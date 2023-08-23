const express = require('express');
const app = express();
const encryption = require("./btoa");
var bodyParser = require('body-parser');
app.use(bodyParser());


app.post('/get_sign', function (req, res) {
    let result = req.body;
    let sign = result.sign;
    console.log(sign);
    result = encryption.aaa(sign);
    res.send(result.toString());
});

app.listen(4001, () => {
    console.log("开启服务，端口 4001")
});
