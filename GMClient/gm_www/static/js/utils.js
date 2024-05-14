function format_time(ts) {
	if (ts == 0) {
		return ""
	}
    var d = new Date(ts * 1000);
    return d.getFullYear() + "-" + (d.getMonth()+1) + "-" + d.getDate() + " " + d.getHours() + ":" + d.getMinutes() + ":"+d.getSeconds()
}