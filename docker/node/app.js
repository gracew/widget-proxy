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
  const customLogic = require(customLogicDir + file);
  const fileNoExt = file.substring(0, file.length - 3);
  app.post("/" + fileNoExt, (req, res) => {
    const output = customLogic(req.body);
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(output));
  });
});

app.get("/ping", (req, res) => res.end());
app.listen(port, () => console.log(`Listening on port ${port}`));
