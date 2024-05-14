/**
 * @author miskice
 * 
 * to gen-survery.html
 */
var _pathUrl = "json/yy.json";
var _defaultChatId = "Newuser";
var sysplat = {1:[],2:[],3:[]};

jQuery(function() {
	bindBtnEvent();	
	/*
	$.getJSON(_pathUrl, function(jsondata){	
		var _defaultChatId = "Newuser";
		showChatAndTable(jsondata, _defaultChatId);
	    //initCharts(jsondata, _defaultChatId);
	    //initTable(jsondata); 
	});
	*/
	// initPlatTable();
	// showPlatList(0);
	// showZoneList(0);
	
	search(_defaultChatId);
});


function initPlatTable() {
	$.getJSON("json/plat.json", function(allplat){
	    var jsonPlat = eval(allplat.data);
	    for(var i=0; i<jsonPlat.length; i++) {
	    	sysplat[jsonPlat[i].system].push(jsonPlat[i].platid);
	    }
	});	
}

function compareObj(keyname) {
	return function(obj1, obj2) {
		if (obj1[keyname] > obj2[keyname]) {
			return 1;
		} else {
			return -1;
		}
	};
}

function showChatAndTable(jsondata, titID) {
	var data = {};
	if (!(jsondata && jsondata.data)) {
		data = {"data":[]};
	} else {
		data = jsondata;
		
		for(var i=0; i<data.data.length; i++) {
			jkeys = ["Gameid","Zoneid","Platid","Daynum","Newuser","Dau","Paynum","Paytoday","Arppu","Arpu"];
			for (var j=0; j<jkeys.length; j++) {
				data.data[i][jkeys[j]] = String(data.data[i][jkeys[j]]);
			}
			jkeys = ["Payrate","Retainedday2","Retainedday3","Retainedday7","Retainedday30"];
			for (var j=0; j<jkeys.length; j++) {
				if (!isNaN(data.data[i][jkeys[j]])) {
					data.data[i][jkeys[j]] = String((Number(data.data[i][jkeys[j]])*100).toFixed(2))+"%";
				}
			}
		}
	}
	data.data.sort(compareObj("Daynum"));
	initCharts(data, titID);
	initTable(data); 
}

function initTable(jsonData){ 		
		$('#sample-table-2').DataTable( {
		"bDestroy":true,		 
		"data":jsonData.data,
		"aoColumns": [
            { "data": "Daynum" },
            { "data": "Newuser" },
            { "data": "Dau" },
            { "data": "Paynum" },
            { "data": "Paytoday" },
            { "data": "Arppu" },
            { "data": "Arpu" },
            { "data": "Payrate" },
			{ "data": "Cashouttoady"},
            { "data": "CashoutRatio" },
            { "data": "CashoutScale" },
            { "data": "Retainedday2" },
            { "data": "Retainedday3" },
            { "data": "Retainedday7" }
        ],
		"order": [[0, 'desc']]
	});		
}

//绑定	
function bindBtnEvent(){		
		//查询
		$('#searchBtn').bind('click',function (){				
			search(_defaultChatId);			
		});
		$("#systemlist").bind('click',function(e){
			changePlatList(e.target.id);
		});
}

//查询
function search(searchid) {
		var tmpdata = $("#reportrange-input").val();
		var stime = Number(tmpdata.substring(0, tmpdata.indexOf('-')));
		var etime = Number(tmpdata.substring(tmpdata.indexOf('-')+1, tmpdata.length));
		// var systype = Number($("#gameSystem").attr("data-id")) || 0;
        //var platid = Number($("#gamePlatform").attr("data-id")) || 0;
        var platid = $("#gamePlat").val();
		var systype = $("#Plattype").val();
        //var zoneid = Number($("#gameZone").attr("data-id")) || 0;
        var zoneid = Number($("#gameZone option:selected").val())||0;

		var gameSystem = $("#lacunch option:selected").val();
		var launchAccount = $("#launchAccountval").val()

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
        var gameid = Number($.cookie("gameid"));
		$.post("/monitor/http", {cmd: "user_daily_data", gameid: gameid, zoneid: zoneid, platlist:platlist.join(","), starttime: stime, endtime: etime,gameSystem:gameSystem,launchAccount:launchAccount}, function(data){
			
			//showChatAndTable(data, searchid);
		    initCharts(data, searchid);
		    
		    initTable(data); 
		}, "json");
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

function initPlatList(allplat, data,systemsid) {
	var jsonPlat = eval(allplat.data);	
	$('#platid').html('');
	plathtml = '<option value="0"></option>';
	// if (!(data && data.data)) {
	// 	//列出所有渠道
	// 	for(var i=0; i<jsonPlat.length; i++){
	// 		plathtml += '<option value="'+jsonPlat[i].platid+'" class="system system'+jsonPlat[i].system+'">'+jsonPlat[i].platname+'</option>';
	// 	}
	// } else {
	// 	for(var i=0; i<jsonPlat.length; i++){
	// 		for(var j=0; j<data.data.length; j++){
	// 			if (jsonPlat[i].platid == data.data[j]){//systemsid==jsonPlat[i].system

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
	$('#platid').html(plathtml);
	// $(plathtml).appendTo('#gamePlatform');
	// $("#gamePlatform").trigger("chosen:updated");	
	// $('#gamePlatform').chosen();
	
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
	
	// showPlatList(gameid);
}

//图表初始化
function initCharts (jsonData, titID){				
		var arrayTime = new Array();
		var arrayItems = new Array();
		//var jsonHb = eval(jsonData.data) || [];
		var jsonHb = jsonData.data || [];
		var chartTitle, chartText, valueSuffix;
		if(titID!=""||titID==undefind||titID==null){
			switch (titID) {
			  case "Newuser": chartTitle ="新增用户";chartText = "新增人数"; valueSuffix = "人";
			    break;
			  case "Dau": chartTitle ="活跃玩家";chartText = "活跃人数"; valueSuffix = "人";
			    break;
			  case "Paynum": chartTitle ="充值玩家";chartText = "充值玩家人数"; valueSuffix = "人";
			    break;
			  case "Paytoday": chartTitle ="充值金额";chartText = "充值金额（元）"; valueSuffix = "元";
			    break;
			  case "Arppu": chartTitle ="ARPPU";chartText = "平均付费玩家收入"; valueSuffix = "元";
			    break;
			  case "Arpu": chartTitle ="ARPU";chartText = "平均玩家收入"; valueSuffix = "元";
			    break;
			  case "Payrate": chartTitle ="付费率";chartText = "付费率"; valueSuffix = "%";
			    break;
			  case "Cashouttoady": chartTitle ="提现金额";chartText = "提现金额"; valueSuffix = "元";
			    break;
			  case "CashoutRatio": chartTitle ="提现率";chartText = "提现率"; valueSuffix = "%";
			    break;
			  case "CashoutScale": chartTitle ="提现比";chartText = "提现比"; valueSuffix = "%";
			    break; 
 			  case "Retainedday2": chartTitle ="次日存留";chartText = "次日存留"; valueSuffix = "%";
			    break;
			  case "Retainedday3": chartTitle ="三日留存";chartText = "三日留存"; valueSuffix = "%";
			    break;
			  case "Retainedday7": chartTitle ="七日留存";chartText = "七日留存"; valueSuffix = "%";
			    break; 
			    
			  default: chartTitle ="新增用户";chartText = "新增人数";valueSuffix = "人";
			}
			
			for(var i=0; i<jsonHb.length; i++) {
			 	arrayTime.push(jsonHb[i].Daynum);
			 	if(parseFloat(jsonHb[i][titID])!= NaN ){
			 		arrayItems.push(parseFloat(jsonHb[i][titID]));//根据传进来的ID变量，取不同的值
			 	}
		 }
		 $('#container').highcharts({
		 	credits: {
			     enabled:false//右下角图标去掉
			},
	    	chart: {
            	type: 'spline'
       		 },
	        title: {
	            text: chartTitle+'图表',
	            x: -20 //center
	        },
	        xAxis: {
				reversed: true,
	            categories: arrayTime
	        },
	        yAxis: {
	            title: {
	                text: chartTitle
	            },
	            plotLines: [{
	                value: 0,
	                width: 1,
	                color: '#808080'
	            }]
	        },
	        tooltip: {
	            valueSuffix: valueSuffix
	        },	        
	        series: [{
	            name: chartText,	            
	            data: arrayItems
	        }]
	    });
		 
		}else{
			$('#container').hide();
			return ;
		}
		
	
}


//图表tab切换
$("#btnChartsGroup .btnCharts").click(function(){
	
	
	$(this).siblings().removeClass('on');
	$(this).addClass('on');
	
	_defaultChatId = $(this).attr('id');	
	search(_defaultChatId);
});
$("#launchAccountval").click(function(){
	$("#launchAccountval").val("");
})