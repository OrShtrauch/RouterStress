FROM python:3.10.6-slim-buster

RUN pip3 install --upgrade pip
RUN pip3 install pandas
RUN pip3 install matplotlib

WORKDIR /var/tmp/stress
RUN mkdir data/

COPY ./plotter.py script/plotter.py

CMD ["python3", "script/plotter.py"]



