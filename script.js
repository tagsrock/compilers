
// string to bytes
String.prototype.getBytes = function () {
  var bytes = [];
  for (var i = 0; i < this.length; ++i) {
    bytes.push(this.charCodeAt(i));
  }
  return bytes;
};

// general framework for ajax calls. 

function new_request_obj(){
    if (window.XMLHttpRequest)
        return new XMLHttpRequest();
    else
        return new ActiveXObject("Microsoft.XMLHTTP");
}

function register_callback(xmlhttp, _func, args){
    xmlhttp.onreadystatechange=function(){
        if (xmlhttp.readyState==4 && xmlhttp.status==200){
		args.unshift(xmlhttp);
		_func.apply(this, args);
        }
    }
}

function make_request(xmlhttp, method, path, async, params){
    xmlhttp.open(method, path, async);
    xmlhttp.setRequestHeader("Content-type", "application/json");
    //xmlhttp.setRequestHeader("Content-length", s.length); // important?
    xmlhttp.send(JSON.stringify(params));
}

function compile_callback(xmlhttp){
   response = JSON.parse(xmlhttp.responseText);   
   console.log(response)
   bytecode = response['bytecode']
   console.log(bytecode)
   document.getElementById("bytecode").innerHTML = "Compiled bytecode: 0x"+bytecode;
}

function compile(){
    code = document.getElementById("code").value;
    codebytes = code.getBytes() 
    console.log(code)
    console.log(codebytes)
    c = [codebytes]
    console.log(c)
    xmlhttp = new_request_obj();
    register_callback(xmlhttp, compile_callback, []);
    make_request(xmlhttp, "POST", "/compile2", true, {"scripts":c});
    return false;
}



