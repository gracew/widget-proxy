FROM mhart/alpine-node:13

RUN npm install express
ADD . /app/
WORKDIR /app

EXPOSE 8080
ENTRYPOINT [ "node", "app.js" ]