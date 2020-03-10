const fs = require("fs");

const express = require("express");
const bodyParser = require("body-parser");

const app = express();
const port = 8080;

app.use(bodyParser.json());

const BEFORE_CREATE = "./customLogic/beforeCreate.js";
const AFTER_CREATE = "./customLogic/afterCreate.js";

if (fs.existsSync(BEFORE_CREATE)) {
  const beforeCreate = require(BEFORE_CREATE);
  app.post("/beforeCreate", (req, res) => {
    const output = beforeCreate(req.body);
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(output));
  });
}

if (fs.existsSync(AFTER_CREATE)) {
  const afterCreate = require(AFTER_CREATE);
  app.post("/afterCreate", (req, res) => {
    const output = afterCreate(req.body);
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify(output));
  });
}

app.listen(port, () => console.log(`Listening on port ${port}`));
