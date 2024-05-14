//菜单
$(document).ready(function() {
	$(".jm-navdrop").hover(function(){
		$(this).addClass("jm-active");
		$(this).children(".jm-ndropbox").stop(true).show();
	},function(){
		$(this).removeClass("jm-active");
		$(this).children(".jm-ndropbox").stop(true).hide();
	});
	$(".jm-ndropbox a").click(function(){
		
		var txt = $(this).text();
		var tId = $(this).attr('id');
		$(this).parent(".jm-ndropbox").prev(".jm-open").html(txt);
		$(this).parent(".jm-ndropbox").prev(".jm-open").attr('id',tId);
		$(this).parent(".jm-ndropbox").stop(true).hide();
		//$(".jm-ndropbox a").css("color","#373737");
		//$(this).css("color","#e37388");
	});

	$("form input").attr("autocomplete","off");
});

//table
$(document).ready(function() {
	$('.jmtable-type01 tbody tr:odd').addClass('even');
	$('.jmtable-type01 tbody tr').hover(
		function() {$(this).addClass('highlight');},
		function() {$(this).removeClass('highlight');}
	);
	$('.jmtable-type01 input[type="checkbox"]:checked').parents('tr').addClass('selected');  
	$('.jmtable-type01 tbody tr').on('click', 
		function() {  
			if ($(this).hasClass('selected')) {  
				$(this).removeClass('selected');  
				$(this).find('input[type="checkbox"]').removeAttr('checked');  
			} else {  
				$(this).addClass('selected');  
				$(this).find('input[type="checkbox"]').attr('checked','checked');
			}  
		}  
	); 
	$(".selectAll").on('click',function(){    
		if(this.checked){    
			$(".jmtable-type01 :checkbox").attr("checked", true);
			$('.check_box').parent().parent('tr').addClass('selected');     
		}else{    
			$(".jmtable-type01 :checkbox").attr("checked", false); 
			$('.check_box').parent().parent('tr').removeClass('selected');
		}    
	});	

	var html = "<option value='0' selected>马甲包</option>";
        html += "<option value='1'>5R5</option>";
        $("#platid").html(html)

});

//tooltip
$(function(){
    var x = -80;  
    var y = -52;
    $(".tipTtile").mouseover(function(e){
        this.myTitle = this.title;
        this.title = "";    
        var tooltip = "<div class='tooltip'><span class='arrow'></span>"+ this.myTitle +"<\/div>";
        $("body").append(tooltip);
        $(".tooltip")
            .css({
                "top": (e.pageY+y) + "px",
                "left": (e.pageX+x)  + "px"
            }).show("fast");
	}).mouseout(function(){        
		this.title = this.myTitle;
		$(".tooltip").remove();
	}).mousemove(function(e){
		$(".tooltip")
		.css({
			"top": (e.pageY+y) + "px",
			"left": (e.pageX+x)  + "px"
		});
	});
})


//tip
$(document).ready(function(){
	$(".datacon").hover(function(){
		$(this).children(".datamsg").toggle(100);
	},function(){
		$(this).children(".datamsg").toggle();
   });
});


//显示弹框 
function showCont(id){
	$("#TB_overlayBG").css({
		display:"block",height:$(document).height()
	});
	$("#" + id).css({
		left:($("body").width()-$("#" + id).width())/2-20+"px",
		top:($(window).height()-$("#" + id).height())/2+$(window).scrollTop()+"px",
		display:"block"
	});
	//$(".scrollbox").tinyscrollbar();
	$(".pop-transform").jqTransform();
}

// 关闭弹框 
function closeCont(id){
	$("#TB_overlayBG").css("display","none");
	$("#" + id).css("display","none");
}

//下拉
$(function(){   
	$(".include_drop").hover(
		function() {
			$(this).children(".droplist").show();
		},
		function() {
			$(this).children(".droplist").hide();
		}
	);
	$(".droplist li").click(function(){   
		$(this).blur();   
		var txt = $(this).text();   
		$(this).parent().prev().val(txt);
		$(".droplist").hide(); 
	}); 
})
$(".droplist li").hover(
	function() {
		$(this).addClass("dhover");
	},
	function() {
		$(this).removeClass("dhover");
	}
)

// 下拉框左右选择
$(document).ready(function(){
	$("#addRight").click(function(){
		$("#select1 option:selected").appendTo("#select2");
	});
	$("#addAllRight").click(function(){
		$("#select1 option").appendTo("#select2");
	});
	$("#addLeft").click(function(){
		$("#select2 option:selected").appendTo("#select1");
	});	
	$("#addAllLeft").click(function(){
		$("#select2 option").appendTo("#select1");
	});	
	$("#select1").dblclick(function(){
		$("#select1 option:selected").appendTo("#select2");	
	});
	$("#select2").dblclick(function(){
		$("#select2 option:selected").appendTo("#select1");	
	});
});




/**
 * 加
 */
$(".g-plus").click(function() {
	var count = $(this).siblings(".ipt_customer_num");
	if(count.val() == "") {
		parent.layer.open({
				type: 0,
				title :'提示信息',
				area: ['300px', ''],
				shadeClose: true, //点击遮罩关闭
				content: '请输入数量！'
			});
	} else if(parseInt(count.val()) < 0) {
		count.val(1);
		return;
	} else {
		count.val(parseInt(count.val()) + 1);
	}
});

/**
 * 减
 */
$(".g-minus").click(function() {
	var count = $(this).siblings(".ipt_customer_num");
	if(count.val() == "") {
		parent.layer.open({
				type: 0,
				title :'提示信息',
				area: ['300px', ''],
				shadeClose: true, //点击遮罩关闭
				content: '请输入数量！'
			});
	} else if(parseInt(count.val()) < 2) {
		count.val(1);
		return;
	} else {
	   count.val(parseInt(count.val()) - 1);	   
	}
});

$(".ipt_customer_num").blur(function() {
	var str = $(this).val();
	if(isNaN(str)){		
		$(this).val(1);
		}else{	
		return;		
		}
	if($(this).val() == "" || parseInt($(this).val()) < 1) {
		$(this).val(1);
	}
});

$(".ipt_customer_zore").blur(function() {
	var str = $(this).val();
	if(isNaN(str)){		
		$(this).val(0);
		}else{	
		return;		
		}
	if($(this).val() == "" || parseInt($(this).val()) < 1) {
		$(this).val(0);
	}
});

/**
 * 添加到选择框
 */
$("#addRewardTobox").click(function(){
	//alert('sssss');
	var _stype = $("#rewardType").val();
	var _stypeInfo = $("#rewardTypeInfo").val();
	var s_typeNum = $("#rewardTypt_num").val();
	var sLiHml = "<li><span>"+_stype+"</span>|<span>"+_stypeInfo+"</span>|<span>数量x"+s_typeNum+"</span><a class='deleteBtn' href='javascript:;' onClick='_sdeleteBtn(this);'>删除</a></li>";	
	$(sLiHml).appendTo("#rewardroBoxUl");
});

//email single
$("#emailAtta_add").click(function(){
	var _stype = $("#attachmenttype option:selected").text();
	var _stypeVal = $("#attachmenttype option:selected").val();
	var _resid = $("#attrresid option:selected").val();
	var _attaName = $("#attaName").val();		
	var _typeNum = $("#emailAtta_num").val();
	var _bindNum = $('.controlBoxs input[name="binding"]:checked ').val()||0;
	var _bindName = "否";

	if(_bindNum == 1 ){
		_bindName = "是";
	}
	
	if(_stypeVal==0){
		dialogTip("请选择物品！");
		return false;
		}else{
			var sLiHml = "<li><span class='s1 t1' data-id='"+_stypeVal+"'>"+_stype+"</span><span class='s1 t2'>"+_resid+"</span><span class='s2 t3'>"+_attaName+"</span><span class='s4 t4'>"+_typeNum+"</span><span class='s4 t5' data-bind='"+_bindNum+"'>"+_bindName+"</span><a class='deleteBtn' href='javascript:;' onClick='_sdeleteBtn(this);'>删除</a></li>"
	
		$(sLiHml).appendTo("#ea-content-ul");
			
	}
	
	//var sLiHml = "<li><span>"+_stype+"</span>|<span>"+_stypeInfo+"</span>|<span>数量x"+s_typeNum+"</span><a class='deleteBtn' href='javascript:;' onClick='_sdeleteBtn(this);'>删除</a></li>";	
	
	
});

function _sdeleteBtn(s){
	$(s).parents("li").slideUp("500", function() {		
		$(this).remove();
	});
}


function dateFormat(date, format){
	date = new Date(date * 1000);  
       var map = {  
           "M": date.getMonth() + 1, //月份   
           "d": date.getDate(), //日   
           "h": date.getHours(), //小时   
           "m": date.getMinutes(), //分   
           "s": date.getSeconds(), //秒   
           "q": Math.floor((date.getMonth() + 3) / 3), //季度   
           "S": date.getMilliseconds() //毫秒   
       };  
       format = format.replace(/([yMdhmsqS])+/g, function(all, t){  
           var v = map[t];  
           if(v !== undefined){  
               if(all.length > 1){  
                   v = '0' + v;  
                   v = v.substr(v.length-2);  
               }  
               return v;  
           }  
           else if(t === 'y'){  
               return (date.getFullYear() + '').substr(4 - all.length);  
           }  
           return all;  
       });  
       return format;  
	}
	
	//提示框
	function showLayer(errortips){
		parent.layer.load({
			type: 0,
			title :'提示信息',
			area: ['300px', ''],
			shadeClose: true, //点击遮罩关闭
			content: errortips
		});
	};
	
function dialogTip(args){
		parent.layer.open({
			type: 0,
			title :'提示信息',
			area: ['300px', ''],
			shadeClose: true, //点击遮罩关闭
			content: args
		});
		}
function choise(){
	var page = $("#choise_page").val();
	if (page <= 0){
		alert("请输入需要跳转的页数")
		return
	}
	listToPage(page)
}
function jumpToPage(id,data){
	var elemet = window.parent.document.getElementById(id);
	var itemid=$(elemet).text()+"_item";

	
	var _ObjtabItems = window.parent.document.getElementsByClassName("add-tab");
	var sArrayTabItem = new Array();
	$(_ObjtabItems).each(function(n) {
		var tabItemsId = $(_ObjtabItems).eq(n).attr('id').replace("_item","");
		sArrayTabItem.push(tabItemsId);
	});
	//如果在下拉列表，删除掉						
	if (sArrayTabItem.indexOf($(elemet).text()) != -1)
	{	
		window.parent.document.getElementById(itemid).remove()
	}


	window.parent.newIframes(id+'.html?'+data , window.parent.document.getElementById(id));
	
}
