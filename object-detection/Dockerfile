FROM python:3.8-slim

WORKDIR /object-detection

COPY . /object-detection

RUN pip install --trusted-host pypi.python.org -r requirements.txt

EXPOSE 6000

CMD ["python","yolov5/app.py"]