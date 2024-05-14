function getadaccount(){
    // var gameid = Number($.cookie("gameid")) || 0;
    // $.post("/monitor/http", {cmd: "launch_channel_list" , gameid : gameid},
    //     function(data){
    //         if (data.retcode == 1) {
    //             // alert(data.retdesc);
    //         } else {
    //             if(data.data==null){
    //                 // alert("加载失败！")
    //                 return
    //             }
    //             var html = "";
    //             for(var i=0 ; i < data.data.length ; i++){
                    
    //                 html += '<option value="'+data.data[i].keywords+'" >'+data.data[i].keywords+'</option>' 
    //             }
               
    //             $("#launchAccount").html(html)
                
    //         }
    //     }, "json");
}
function timezong(){

    let targetTimezone = +3

    let _dif = new Date().getTimezoneOffset()

    let east9time = new Date().getTime() + _dif * 60 * 1000 - (targetTimezone * 60 * 60 * 1000)

    return new Date(east9time)
}

// setInterval(function(){ 
//     var gameid = Number($.cookie("gameid")) || 0;
//     if (gameid != 1001){
//         $.cookie("gameid", 1001);
//     }
//   }, 1000*3);