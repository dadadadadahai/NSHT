var currentid="";
var currentitem="";
var fisrtIfrid="";
var fisrttitemid="";
var itemArray=new Array();

function autoSize(){
    var maxWidth = Math.min(
        document.documentElement["clientWidth"],
        document.body["scrollWidth"], document.documentElement["scrollWidth"],
        document.body["offsetWidth"], document.documentElement["offsetWidth"]
    ); 
    $("#left_menu").css({
		height: ( document.documentElement["clientHeight"] - 70 ) +'px',
	});
	$(".right-frame").css({
		width: ( maxWidth - 170 ) +'px',
		height: ( document.documentElement["clientHeight"] - 70 ) +'px',
	});
	
	$(".jm-subcon").css({
		height: ( document.documentElement["clientHeight"] - 110 ) +'px',
	});
	$(".jm-subcon").tinyscrollbar();

}
//autoSize();
window.onresize = autoSize;

function sLiSide(itemid_s,url){
	var _iwidth = $('.jm-menubar').width()-170;	
	var sLi = $('#jmMenulist li');
	var sLiWidth = 0;	
	if (sLi.length == 0){		
		return;
		}else{
			var _fristLiWidth = $(sLi).eq(0).width();	
			$(sLi).each(function(i) {
				sLiWidth += ($(sLi).eq(i).width()+10);
			
				if(sLiWidth>(_iwidth-_fristLiWidth-40)){
					$(".slideIconBox_a").show();
					var _sItemId = $(sLi).eq(i).attr("id");
					var _sListItemId = _sItemId.replace("_item","");//当前的id值

				
					var itemid_s_ht = itemid_s.replace("_item","");//获取传过来的ID值					
					var _sListHml = "<div class='menuListCloseRT' id='"+_sListItemId+"_sList' data-url='"+url+"'>"+_sListItemId+"</div>";									
					var _objListItemId = $(".menuListCloseRT");
					var sArrayItem = new Array();
					
					//判断是不是在下拉列表中
					$(_objListItemId).each(function(k) {						
						//获取下拉列表中的所有值
						var sslistId = $(_objListItemId).eq(k).attr('id').replace("_sList","");						
                        sArrayItem.push(sslistId);						
                    });	
					//如果在下拉列表，删除掉						
					if (sArrayItem.indexOf(itemid_s_ht) != -1)
					{	
						$("#"+itemid_s_ht+"_sList").remove();	
					}
					
					//判断是否在当前的列表中
					var _ObjtabItems = $(".add-tab");
					var sArrayTabItem = new Array();
					$(_ObjtabItems).each(function(n) {
                        var tabItemsId = $(_ObjtabItems).eq(n).attr('id').replace("_item","");
						sArrayTabItem.push(tabItemsId);
                    });
                    //如果不在tab里，则删掉最后一个					
					if (sArrayTabItem.indexOf(itemid_s_ht) == -1)
						{	$(_sListHml).appendTo("#menuListc");			
							$(sLi).eq(i).remove();		
						}						
					//$(sLi).eq(i).remove();					
				}else{
					$(".slideIconBox_a").hide();
					 $("#sIBox-info").hide();
					}
            }); 			
		}			
	}
	

//关闭全部列表
$(".menuClose").click(function(){
	var sTabList = $('#jmMenulist li');
	var sIframeList = $('#contain iframe');
	var shideList = $('.menuListc .menuListClose');
	$(sTabList).not($(sTabList).first()).remove();
	$(shideList).remove();
	$(".menuListc .menuListCloseRT").remove(); 
	$("#sIBox-info").slideUp();
	$(".slideIconBox_a").fadeOut();	
	$(sIframeList).not($(sIframeList).first()).remove();	
	$('.jm-dropmenu a').first('a')[0].click();	
	$('.bg_zindex').hide();	
});


//右侧隐藏列表选中点击事件
$(".menuListCloseRT").on("click",function(){
	var dataUrl = $(this).attr('data-url');
	newIframes(dataUrl,this);
	$(this).remove();
	$('.bg_zindex').hide();
	});

//弹出导航列表
$(".slideIconBox_a").click(function(){
	event.stopPropagation();	
	if($("#sIBox-info").is(":hidden")){		
		$("#sIBox-info").slideDown();
		$('.bg_zindex').show();			
		}else{
			$("#sIBox-info").slideUp();
			$('.bg_zindex').hide();
		}
	})

$(window).bind('click',function(){			
	if($("#sIBox-info").is(":visible")){
		$("#sIBox-info").slideUp();
	}else {				
		return false;
	}
	$('.bg_zindex').hide();				
});

$(".bg_zindex").bind('click',function(){			
	if($("#sIBox-info").is(":visible")){
		$("#sIBox-info").slideUp();
	}else {				
		return false;
	}
	$(this).hide();			
});
setInterval(function(){ 
	var gameid = $.cookie("gameid")
	var zoneid = $.cookie("zoneid")
	if(gameid == null || zoneid == null){
		$.cookie("gameid", 1000);
		$.cookie("zoneid", 1);
	}
   
  }, 1000*3);
function newIframes(url,obj){

	var sgameId = $.cookie("gameid");
	var szoneId = $.cookie("zoneid");
	/*var d = dialog({
		width: 250,
		title: '提示',
		content: '请先选择游戏列表以及游戏区服！',
		button: [{value: '确定'}],
	});*/

	var stid = $(obj).attr('id');
	var itemid=$(obj).text()+"_item";
	var ifrid=$(obj).text()+"_frame";
	
	
	var isFtistId = $('.jm-dropmenu a').first('a').attr('id');
	

	//判断是否选择游戏列表
	if(isFtistId==stid){
			//d.showModal();
			}
	if(sgameId&&szoneId||(isFtistId==stid || stid == "user_info_search_kf")){
		
		sLiSide(itemid,url);
		
		if(!(typeof($("#"+itemid).attr("id"))=='undefined' || typeof($("#"+itemid).attr("id"))=='' )){
			
			if($(obj).text() == "玩家查找"){
				return
			}
			$("#"+currentid).css("display","none");
			$("#"+ifrid).css("display","block");
			$("#"+currentitem).attr("class","add-tab");
			currentid=ifrid;
			currentitem=itemid;
			$("#"+currentitem).attr("class","add-tab tab-local");
			return ;
		}
		
		

		var ifr="<iframe id='" + ifrid + "' class='right-frame' frameborder='0' scrolling='auto' allowTransparency='true'></iframe>"
		var item="";
		
		if($("#jmMenulist").html()==""){
		  item="<li id='" + itemid + "' class='add-tab tab-local'><a href='#'>"+$(obj).text()+"</a></li>";
		  currentid=ifrid;
		  currentitem=itemid;
		  fisrtIfrid=ifrid;
		  fisrttitemid=itemid;
		}else{
		  item="<li id='" + itemid + "' class='add-tab tab-local'><a href='#'><b class='btn-close'></b><span>"+$(obj).text()+"</span></a></li>"
		}
		itemArray.push(itemid);

		$("#"+currentid).css("display","none");
		$("#contain").append(ifr);
		$("#jmMenulist").append(item);
		$("#"+ifrid).attr("src", url)
		$("#"+currentitem).attr("class","add-tab");

		autoSize(ifrid);
		currentid=ifrid;
		currentitem=itemid;
		$("#"+currentitem).attr("class","add-tab tab-local");

		$(".btn-close").click(function() {
			$(this).parent().parent().css("display","none");
			var ifrobj=$(this).parent().parent().attr("id").replace("_item","_frame");
			$("#"+$(this).parent().parent().attr("id")).remove();
			$("#"+ifrobj).attr("src", null);
			$("#"+ifrobj).remove();

			var _objListSize = $(".menuListCloseRT");
			
			if(_objListSize.length>0){
				var _objListSize_item = $(_objListSize).eq(0).attr("id").replace("_sList","");
				$(_objListSize).eq(0).remove();


				}else{

				var itemTempArray=new Array();
				var open=false;
				for(var i=itemArray.length;i>0;i--){
				var tempItem=itemArray.pop();
				  if(!(typeof($("#"+tempItem).attr("id"))=='undefined' || typeof($("#"+tempItem).attr("id"))=='')){
					itemTempArray.push(tempItem);
					if(!open){
						currentid=tempItem.replace("_item","_frame");
						currentitem=tempItem;
						$("#"+tempItem.replace("_item","_frame")).css("display","block");
						$("#"+tempItem).attr("class","add-tab tab-local");
						open=true;
					}
				  }
				}
				for(var i=itemTempArray.length;i>0;i--){
					itemArray.push(itemTempArray.pop());
				}
				//sLiSide(itemid,url);
			}
		});

		$(".add-tab").click(function() {
			var ifrobj=$(this).attr("id").replace("_item","_frame");
			$("#"+currentid).css("display","none");
			$("#"+currentitem).attr("class","add-tab");
			currentid=ifrobj;
			currentitem=$(this).attr("id");
			$("#"+ifrobj).css("display","block");
			$("#"+currentitem).attr("class","add-tab tab-local");
			itemArray.push(currentitem);
		});
		console.log(itemArray);

	}else{
		//d.showModal();
		layer.open({
			type: 0,
			title :'提示信息',
			area: ['300px', ''],
			shadeClose: true, //点击遮罩关闭
			content: '请先选择游戏列表以及游戏区服！'
		});
	}
}


//选中游戏时改变区服列表
function ongamechange(e){
	$.cookie("gameid", e.target.id);
	$.cookie("zoneid", null);
	$("#zone-list a").css("display","none");
	$(".game"+e.target.id).css("display","block");
}
//区服改变时添加cookie
function onzonechange(e) {
	$.cookie("zoneid", e.target.id);
}

/*
function attachmenu(){
	var txt = $(this).text();
	var htm = "<li class='leclose'><a href='#'><b></b><span>"+txt+"</span></a></li>";
	$(window.parent.document.getElementById("jmMenulist")).append(htm)
	$("#jmMenulist .leclose").click(function() {
		$(this).remove();
	})
}*/


$(document).ready(function(){
	$.cookie("zoneid", 1);
	$.cookie("gameid", 1000);
	$.cookie("_xsrf",null);
	$("#game-list a").click(ongamechange);
	$("#zone-list a").click(onzonechange);
	$(".jm-subcon").tinyscrollbar();
	
	$(".jm-submenu .jm-first").click(function() {
		var sgameId = $.cookie("gameid");
		var szoneId = $.cookie("zoneid");		
		
		if(sgameId&&szoneId){			
			//$(this).toggleClass("jm-darrow");
			$('.jm-first').addClass("jm-darrow");
			if($(this).hasClass("jm-darrow")){
				$(this).removeClass("jm-darrow");
				}else{
				$(this).addClass("jm-darrow");
			}			
			$(".jm-dropmenu").slideUp();
			$('.jm-first').removeClass('cur');
			$(this).next(".jm-dropmenu").stop(true).slideToggle(function(){				
				$(".jm-subcon").tinyscrollbar();
				});	
			$(this).addClass('cur');
			$(".jm-subcon").tinyscrollbar();
		}else{
			layer.open({
				type: 0,
				title :'提示信息',
				area: ['300px', ''],
				shadeClose: true, //点击遮罩关闭
				content: '请先选择游戏列表以及游戏区服！'
			});
			
			}
		
	});
	
	$('.jm-dropmenu a').click(function(){
		$(".jm-subcon").tinyscrollbar();
		}	
	)
});
