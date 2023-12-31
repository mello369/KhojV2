FROM python:3.8-slim

WORKDIR /yolov5-app

COPY . /yolov5-app

RUN pip install --trusted-host pypi.python.org -r requirements.txt

EXPOSE 6000

CMD ["python","yolov5/app.py"]