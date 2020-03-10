FROM python:3.8-slim

RUN pip install Flask==1.1.1
ADD . /app/
WORKDIR /app

ENV FLASK_APP app.py
EXPOSE 8080
ENTRYPOINT [ "flask", "run", "--host", "0.0.0.0", "--port", "8080" ]