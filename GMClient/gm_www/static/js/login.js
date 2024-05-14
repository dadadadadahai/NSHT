
function check_login(){
	return ($.cookie("token") || "") != "";
}

function clearCookie(){
    var keys = document.cookie.match(/[^ =;]+(?=\=)/g);
    if (keys) {
        for (var i = keys.length; i--; )
            document.cookie = keys[i] + '=0;expires=' + new Date(0).toUTCString()
    }    
}

function login(){
	var username = $("#fusername").val();
	var password = $("#fpassword").val();
	var xsrftoken = $("#_xsrf").val();
	var onLoadTip;
	$("#fpassword").val("");
	if(username == "" || password == "" || xsrftoken == "") {
		$("#ferror").text("请输入用户名以及密码！").show();
		return;
		}else{
			$("#ferror").text('');
			}
	var remeber = $("#fremember").attr("checked") == "checked" ? "true" : "";
	/*$.post("/gm/http", {cmd: "gmlogin", username: username, password: password, remeber: remeber, _xsrf: xsrftoken}, function(data){
		$("#fpassword").val("");
		if (data.retcode == "0") {
			$.cookie("username", username, {expires: null});
			top.location = "/home.html";
		} else {			
			$("#ferror").text("Error:"+data.retdesc).show();
		}
	}, "json");*/
	$.ajax({
		type:"post",
		url:"/gm/http",
		data: {cmd: "gmlogin", username: username, password: password, remeber: remeber, _xsrf: xsrftoken},
		dataType:"json",
		beforeSend: function(){
			onLoadTip = layer.load(1, {
				shade: [0.1,'#fff'] //0.1透明度的白色背景
			});
		},
		success: function(data){
			
			$("#fpassword").val("");
			if (data.retcode == "0") {
				$.cookie("username", username, {expires: null});
				layer.close(onLoadTip);
				top.location = "/home.html";
			} else {
				layer.close(onLoadTip);			
				//$("#ferror").text("Error:"+data.retdesc).show();
				$("#ferror").text("用户名密码错误，请重新输入！").show();
			}	
		}
	});
	
}

function logout(){
	$.post("/logout",{},function(){});
	clearCookie();
	top.location = "/login.html";
}

$(document).ready(function(){
	$("#fpassword").keypress(function(e){ if (e.which == 13) login();});
	$("#flogin").click(login);
});



