# 补环境

1.调试环境

2.在环境框架中直接运行某js文件，补修改原js文件

## 自己写一个浏览器

1.BOM 浏览器实现的代码历史记录浏览器版本信息之类的

2.DOM 根据id取元素

3.网络请求 xhr jsonp jsonp_3213123({ayf:1})

4.跨窗口worker

目的： 运行环境

难点：如何找到缺少的那些环境，如何很好的实现环境代码

伪造：我给你啥，你能正确返回啥可以了

## vm2 纯净的v8环境

脱离了nodejs的v8环境，防止检测到nodejs环境

https://github.com/patriksimek/vm2
```javascript
var fs = require('fs');
const {VM} = require('vm2');
const vm = new VM();
var data = fs.readFileSync('./src/code.js', 'utf8')
vm.run(data)
```

调试沙盒代码
```javascript
const {VM, VMScript} = require('vm2');
const fs = require('fs');
// 运行的code代码
const file = `${__dirname}/code.js`;

// 需要补的window环境
const windowfile = `${__dirname}/window.js`;
const vm = new VM();

//VMScript调试
const script = new VMScript(fs.readFileSync(windowfile)+fs.readFileSync(file), "我正在调试的代码");
vm.run(script);
```
