FROM python:3.10

LABEL MAINTAINER 'janog-netcon'

WORKDIR /app

COPY . .

RUN pip install --no-cache-dir -r requirements.txt

ENTRYPOINT ["/app/coordinate.py"]
