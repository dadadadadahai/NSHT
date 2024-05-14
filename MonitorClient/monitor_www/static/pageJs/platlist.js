var sysplat = {1:[],2:[],3:[]};

function initPlatTable() {
	$.getJSON("json/plat.json", function(allplat){
	    var jsonPlat = eval(allplat.data);
	    for(var i=0; i<jsonPlat.length; i++) {
	    	sysplat[jsonPlat[i].system].push(jsonPlat[i].platid);
	    }
	});
}

function getPlatlist(systype, platid){

	var platlist = [];

	var arr = [[1,2,5,6,8], [101,102,105,106,108],[1001,1002,1005,1006,1008]];
	if(platid == null ){

		if (systype == null) {

			platlist = [];
		}else{
			$.each(systype, function(index, value) {

				$.each(arr[value-1], function(ind, val) {

					platlist.push(parseInt(val));
				
				});
				
			});
		}
	}else{
		
		$.each(platid, function(index, value) {
			if(systype == null){
				platlist.push(parseInt(value));
				platlist.push(parseInt(value)+100);
				platlist.push(parseInt(value)+1000);
			}else{
				$.each(systype, function(ind, val) {
					if (val == 2) {
						platlist.push(parseInt(value)+100);
						
					}else if(val == 3){
						platlist.push(parseInt(value)  + 1000);
						
					}else{
						platlist.push(parseInt(value));
					}
				});
			}
		});
	}
	
    return platlist;
}

//获取游戏接入的渠道ID，用于显示渠道列表
function showPlatList(gameid) {
	var systemsid =  Number($("#gameSystem").attr("data-id")) || 0;
	if (gameid == 0) {
		gameid = Number($.cookie("gameid"));
	}
	// if (gameid) {
	// 	$.post("/monitor/http", {cmd: "game_plat_list", gameid:gameid}, 
	//         function(data){
	// 			$.getJSON("json/plat.json", function(allplat){	
	// 			    initPlatList(allplat, data,systemsid);
	// 			});	
	//         }, "json");
	// } else {
		$.getJSON("json/plat.json", function(allplat){	
			    initPlatList(allplat, null,systemsid);
			});	
	// }
}

function initPlatList(allplat, data,systemsid) {
	var jsonPlat = eval(allplat.data);
	$('#gamePlatform').html('');
	plathtml = '<option value="0"></option>';
	// if (!(data && data.data)) {
	// 	//列出所有渠道
	// 	for(var i=0; i<jsonPlat.length; i++){
	// 		plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'">'+jsonPlat[i].platname+'</option>';
	// 	}
	// } else {
	// 	for(var i=0; i<jsonPlat.length; i++){
	// 		for(var j=0; j<data.data.length; j++){
	// 			if (jsonPlat[i].platid == data.data[j]){
	// 				if(systemsid==0){
	// 					plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'">'+jsonPlat[i].platname+'</option>';
	// 					break;
	// 				}
	// 				if (systemsid==jsonPlat[i].system){
	// 					plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'">'+jsonPlat[i].platname+'</option>';
	// 					break;
	// 				}
	// 			}
	// 		}
	// 	}		
	// }
	for(var i=0; i<jsonPlat.length; i++){
		if(jsonPlat[i].selected == 1){
			plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'" selected>'+jsonPlat[i].platname+'</option>';
		}else{
			plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'">'+jsonPlat[i].platname+'</option>';
		}
		
	}
	$('#gamePlatform').html(plathtml);
	//$(plathtml).appendTo('#gamePlatform');
	$("#gamePlatform").trigger("chosen:updated");	
	$('#gamePlatform').chosen();
	

}


function clickToggle(targetclass) {
	$("."+targetclass).on("click",function(){
		var oTxt = $(this).text();
		var data_id = $(this).attr("data-id");
		if(oTxt == null || oTxt == "" ||oTxt == "undefind"){
			return;
		}			
		$(this).parent("li").parent("ul").siblings(".dopdownTxt").children(".filter-option").text(oTxt);
		$(this).parent("li").parent("ul").siblings(".dopdownTxt").children(".filter-option").attr("data-id", data_id);
		$(this).parent("li").parent("ul").siblings(".dopdownTxt").children("input[type='hidden']").text(oTxt);
		$(this).parent("li").parent("ul").siblings(".dopdownTxt").children("input[type='hidden']").attr("data-id", data_id);
	});
}

function showZoneList(gameid) {
	if (gameid == 0) {
		gameid = Number($.cookie("gameid"));
	}
	if (gameid) {
		$.post("/monitor/http", {cmd: "game_zone_list", gameid:gameid}, 
	        function(data){
	        	if (data && data.data) {
	        		zonehtml = '';
	        		for(var i=0; i<data.data.length; i++) {
	        			var zonedata = data.data[i];
	        			zonehtml += '<option value="'+zonedata.zoneid+'">'+zonedata.zoneid+'</option>';	        			
	        		}	        		
	        		$(zonehtml).appendTo('#gameZone');
	        		$('#gameZone').chosen();
	        	}
	        }, "json");
	}
}

//选择系统时，显示相应的渠道
function changePlatList(systype) {
	// if (systype != 0) {
	// 	$(".system").hide();
	// 	$(".system"+systype).show();
	// } else {
	// 	$(".system").show();
	// }
	gameid = Number($.cookie("gameid"));
	showPlatList(gameid);
}

function compareObj(sorted_keys) {
	return function(obj1, obj2) {
		if (obj1[sorted_keys[0]] > obj2[sorted_keys[0]]) {
			return 1;
		} else if (obj1[sorted_keys[0]] < obj2[sorted_keys[0]])  {
			return -1;
		} else if (sorted_keys.length>1){
			if (obj1[sorted_keys[1]] > obj2[sorted_keys[1]]) {
				return 1;
			}
		} 
		return -1;
	}

}

function showTable(jsondata, keys, sorted_keys, contentclass) {
	var data = {}
	if (!(jsondata && jsondata.data)) {
		data = {"data":[]}
	} else {
		data = jsondata
	}
	data.data.sort(compareObj(sorted_keys));
	contenthtml = ""
	for(var i=0; i<data.data.length; i++) {
		curdata = data.data[i]
		contenthtml += "<tr>";
		for(var j=0; j<keys.length; j++) {
			contenthtml += "<td>"+curdata[keys[j]]+"</td>";
		}
		contenthtml += "</tr>";
	}
	$("."+contentclass+" tbody").html(contenthtml);
}

function formatDate(d) {
	if (isNaN(d.getDate())) {
		d = new Date();
	}
	var str = "";
	str += d.getFullYear();
	str += d.getMonth()>=9?d.getMonth()+1:'0'+(d.getMonth()+1);
	str += d.getDate()>9?d.getDate():'0'+d.getDate();
	return Number(str);
}

function yearMonthDay(d, month, day) {
	if(isNaN(d.getDate())){
		d = new Date();
	}
	d.setDate(d.getDate()+day);
	d.setMonth(d.getMonth()+month);
	var str = "";
	str += d.getFullYear();
	str += d.getMonth()>=9?d.getMonth()+1:'0'+(d.getMonth()+1);
	str += d.getDate()>9?d.getDate():'0'+d.getDate();
	return Number(str);

}
function getadaccount(){
    var gameid = Number($.cookie("gameid")) || 0;
    $.post("/monitor/http", {cmd: "launch_channel_list" , gameid : gameid},
        function(data){
            if (data.retcode == 1) {
                // alert(data.retdesc);
            } else {
                if(data.data==null){
                    // alert("加载失败！")
                    return
                }
                var html = "";
                for(var i=0 ; i < data.data.length ; i++){
                    
                    html += '<option value="'+data.data[i].keywords+'" >'+data.data[i].keywords+'</option>' 
                }
               
                $("#launchAccount").html(html)
                
            }
        }, "json");
}
jQuery(function(){
	// initPlatTable();
    // showPlatList(0);
    // showZoneList(0);
	// getadaccount()
    $("#systemlist").bind('click',function(e){
		changePlatList(e.target.id);
	});
});
function timezong(){

    let targetTimezone = +3

    let _dif = new Date().getTimezoneOffset()

    let east9time = new Date().getTime() + _dif * 60 * 1000 - (targetTimezone * 60 * 60 * 1000)

    return new Date(east9time)
}
$("#launchAccountval").click(function(){
	$("#launchAccountval").val("");
})