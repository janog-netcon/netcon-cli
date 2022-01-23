#!/usr/bin/env python3

"""
処理の流れ
- GCPのinstance listを取得する
- vmdbに登録されているVMのインスタンスのリストを取得する
- vmdbに登録されているが、GCPのinstance listに存在しないVMをピックアップする(差分を取る)
- 対象のインスタンスをスコアサーバーから削除する
// ProblemEnvironment.Name が instance name になる
---
前コンテストでの対応
- 作成に失敗したVMはReady状態にならずにスコアサーバに残り続ける
  - Readyにはならないので参加者に割り当てられることはない
  - ただし、問題に対するインスタンス数のquotaに引っかかってしまい、
    ユーザに割り当てられる問題VMが足りなくなってしまう
- スコアサーバ経由でdeleteするのではなく、DBをいじってレコードを消していた？
"""


from sys import argv
import argparse
import collections
import json
import requests
import subprocess

import googleapiclient.discovery


def run_cmd(args):
    res = subprocess.run(args, stdout=subprocess.PIPE)
    return res


def get_instances_from_gcp__gcloud():
    cmd = ["gcloud", "compute", "instances", "list", "--format=json"]
    res = run_cmd(cmd)
    instances = json.loads(res.stdout)

    instance_names = []

    for instance in instances:
        instance_names.append(instance["name"])

    return instance_names


# https://cloud.google.com/compute/docs/tutorials/python-guide?hl=ja#listinginstances
def get_instances_from_gcp(project, zones):
    compute = googleapiclient.discovery.build("compute", "v1")

    instance_names = []
    for zone in zones:
        result = compute.instances().list(project=project, zone=zone).execute()
        if "items" in result:
            for item in result["items"]:
                instance_names.append(item["name"])


def get_instances_from_vmdb(endpoint):
    url = endpoint + "/problem-environments"

    try:
        res = requests.get(url)
    except Exception:
        print("[WARNING] vmdb requests.get error")
        return []

    if res.status_code != 200:
        print(
            "[WARNING] vmdb/problem-environments status_code not 200 -> {}".format(
                res.status_code
            )
        )
        raise Exception("status code not 200")

    # instances = json.loads(res.json())
    instances = res.json()
    # print(instances)

    instance_names = []

    for instance in instances:
        instance_names.append(instance["name"])

    return instance_names


def filter_lost_instances(gcp_instances, vmdb_instances):
    """
    それっぽい関数があってもいいはずなんだけど見つからないのでごり押し
    """
    instances = []
    instances.extend(gcp_instances)
    instances.extend(vmdb_instances)

    counted_instances = collections.Counter(instances)

    filtered_instances = []
    for instance, count in counted_instances.items():
        if count == 1:
            filtered_instances.append(instance)

    # GCPのみに存在しているインスタンス名
    filtered_gcp_only_instances = []
    for instance in filtered_instances:
        if instance in gcp_instances:
            filtered_gcp_only_instances.append(instance)

    # vmdbのみに存在しているインスタンス名
    filtered_vmdb_only_instances = []
    for instance in filtered_instances:
        if instance in vmdb_instances:
            filtered_vmdb_only_instances.append(instance)

    return {
        "all": filtered_instances,
        "gcp_only": filtered_gcp_only_instances,
        "vmdb_only": filtered_vmdb_only_instances,
    }


def delete_lost_instances(vmdb_endpoint, lost_instances):
    for instance_name in lost_instances:
        url = vmdb_endpoint + "/problem-environments/" + instance_name
        try:
            res = requests.delete(url)
        except Exception:
            print("[WARNING] delete_lost_instance: requrests.delete error")
        if res.status_code != 204:
            print(
                "[WARNING] delete_lost_instance: status code not 204 -> {}: target instance name {}".format(
                    res.status_code, instance_name
                )
            )
        print("[INFO] vmdb instance deleted: {}".format(instance_name))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--dry_run", default=False)
    parser.add_argument("--vmdb_url", default="http://localhost:8905")
    args = parser.parse_args()

    vmdb_endpoint = args.vmdb_url

    # project = ""
    # zones = ["asia-northeast1-a", "asia-northeast1-b", "asia-northeast1-c"]

    gcp_instances = get_instances_from_gcp__gcloud()
    # gcp_instances = get_instances_from_gcp(project, zones)
    # gcp_instances = ["image-aoi-1"]

    vmdb_instances = get_instances_from_vmdb(vmdb_endpoint)
    # vmdb_instances = ["image-aoi-1", "image-nao-1"]

    lost_instances = filter_lost_instances(gcp_instances, vmdb_instances)

    print("======")
    print("gcp instances list: {}".format(gcp_instances))
    print("---")
    print("vmdb instances list: {}".format(vmdb_instances))
    print("---")
    print("gcp_only exists instances: {}".format(lost_instances["gcp_only"]))
    print("---")
    print(
        "vmdb_only exists instances (delete target): {}".format(
            lost_instances["vmdb_only"]
        )
    )
    print("======")

    if not args.dry_run:
        delete_lost_instances(vmdb_endpoint, lost_instances["vmdb_only"])


if __name__ == "__main__":
    main()
