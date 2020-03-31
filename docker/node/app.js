const fs = require("fs");

const express = require("express");
const bodyParser = require("body-parser");

const app = express();
const port = 8080;

app.use(bodyParser.json());

const customLogicDir = "./customLogic/";
const files = fs
  .readdirSync(customLogicDir)
  .filter(file => file.endsWith(".js"));
console.log("found files " + files);

files.forEach(file => {
  const fileNoExt = file.substring(0, file.length - 3);
  const customLogic = require(customLogicDir + file);
  app.post("/" + fileNoExt, (req, res) => {
    const output = customLogic(req.body);
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(output));
  });
});

app.get("/ping", (req, res) => res.end("pong"));
app.listen(port, () => console.log(`Listening on port ${port}`));
