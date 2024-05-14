/**
 * @author miskice
 *
 * to gen-survery.html
 */
var _pathUrl = "json/kk.json";
var _defaultChatId = "Newuser";
var sysplat = {1:[],2:[],3:[]};

jQuery(function() {
	bindBtnEvent();
	// initPlatTable();
	// showPlatList(0);
	// showZoneList(0);
	// getadaccount()
	$("#gamePlat , #Plattype").multipleSelect({
		placeholder: "全部",
		selectAllText:'全部',
		allSelected:'全部'
	});

	//select全部选中
	// $("#gamePlat  ,  #Plattype").multipleSelect('checkAll');		
	$('#timedate').datetimepicker({
		timepicker:false,
		format: "yyyy-mm-dd",
		autoclose: true,
		minView:2,
		maxDateNow: true,
		pickerPosition: "bottom-left"
	});
	
});

function initPlatTable() {
	$.getJSON("json/plat.json", function(allplat){
	    var jsonPlat = eval(allplat.data);
	    for(var i=0; i<jsonPlat.length; i++) {
	    	sysplat[jsonPlat[i].system].push(jsonPlat[i].platid);
	    }
	});
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

function initTable(jsonData){
		$('#sample-table-2').DataTable( {
		"data":jsonData.data,
		// "aaSorting":[[0,"desc"]],
		"aoColumns": [
            { "data": "Recordtime" },
            { "data": "Totaluser" },
            { "data": "Newuser" },
			{ "data": "Olduser" },
            { "data": "Dau" },
            { "data": "Paynum" },
            { "data": "Paytoday" },
            { "data": "Arppu" },
            { "data": "Arpu" },
            { "data": "Payrate" },
			{ "data": "Cashoutnum" },
            { "data": "Cashouttoday" },
            { "data": "CashoutRatio" },
            { "data": "CashoutScale" },
            // { "data": "Day2" },
           
        ],
		"order": [[0, 'desc']]

	});

	}

//绑定
function bindBtnEvent()
	{
		//查询
		$('#searchBtn').bind('click',function (){
			search();
		});
		$(".systemlist").bind('click',function(e){
			changePlatList(e.target.id);
		});

	}

function showChatAndTable(jsondata) {
	var data = {}
	if (!(jsondata && jsondata.data)) {
		data = {"data":[]}
	} else {
		data = jsondata
		// for(var i=0; i<data.data.length; i++) {
		// 	var d = new Date(Number(data.data[i]["Recordtime"])*1000)
		// 	data.data[i]["Recordtime"] = d.toTimeString().substring(0,5);
		// }
	}
	initTable(data);
}

//查询
function search (){
	// var systype = Number($("#gameSystem").attr("data-id")) || 0;
	var platid = $("#gamePlat").val();
	var systype = $("#Plattype").val();
    var zoneid = Number($("#gameZone option:selected").val())||0;
	var time1 = $("#timedate").val();

	var reg = new RegExp("-" , "g")
	var time = time1.replace(reg , "")
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
	var gameSystem =$("#lacunch option:selected").val();
	var launchAccount = $("#launchAccountval").val();
	var usertype = $("#usertype").val();

	$.post("/monitor/http", {cmd: "user_realtime_data", gameid: gameid, zoneid: zoneid, platlist:platlist.join(","), gameSystem:gameSystem ,launchAccount:launchAccount,time:time,usertype:usertype }, function(data){
		showChatAndTable(data);
	}, "json");
}

//图表初始化
function initChartsDetails (jsonData,titID){
		var arrayTime = new Array();
		var arrayItems = new Array();
		var arrayDetails = new Array();
		var jsonHb = eval(jsonData.data);
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
			  case "Day2": chartTitle ="次日留存";chartText = "次日留存"; valueSuffix = "%";
			    break;
			  case "Ltv": chartTitle ="实时LTV";chartText = "实时LTV"; valueSuffix = "元";
			    break;

			  default: chartTitle ="新增用户";chartText = "新增人数";valueSuffix = "人";
			}

			for(var i=0; i<jsonHb.length; i++)   {
			 	arrayTime.push(jsonHb[i].ss_time);
			 	if(parseFloat(jsonHb[i][titID])!= NaN ){
			 	arrayItems.push(parseFloat(jsonHb[i][titID]));//根据传进来的ID变量，取不同的值
			 	arrayDetails.push(parseFloat(jsonHb[i].act_player));
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
	        subtitle: {
	            text: '来源: 乐享畅游',
	            x: -20
	        },
	        xAxis: {
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
	            name: "今天",
	            marker: {
	            	enabled: false,
                    symbol: 'circle'//点形状
                },
                color:"#dcad77",
	            data: arrayItems
	        },{
	            name: "昨天",
	            marker: {
	            	enabled: false,
                    symbol: 'circle'//点形状

                },
	            color:'#23b7e5',
	            data: arrayDetails
	        }]
	    });

		}else{
			$('#container').hide();
			return ;
		}

}
$("#launchAccountval").click(function(){
	$("#launchAccountval").val("");
})

