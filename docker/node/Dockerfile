FROM mhart/alpine-node:13

ADD . /app/
WORKDIR /app
RUN npm install

EXPOSE 8080
ENTRYPOINT [ "node", "app.js" ]