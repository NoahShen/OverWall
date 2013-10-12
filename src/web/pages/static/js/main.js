var limit = 20;
var playStatusTimer;
$(function() {
	$("#refresh").click(function(){
		showNews(true)
	});
	
	$("#play").click(function(){
		$.ajax({
			url: "/news/play",
			cache: false
		});
	});

	$("#stop").click(function(){
		$.ajax({
			url: "/news/stop",
			cache: false
		});
	});

	$("#next").click(function(){
		$.ajax({
			url: "/news/next",
			cache: false
		});
	});
	showNews(false);
});

function showNews(refresh) {
	showLoading()
	var url = '/news/getnews?limit=' + limit
	if (refresh) {
		url += "&refresh=1"
	}
	$.getJSON(url, function(result) {
		if (result.result == "success") {
			var list = $("#one [data-role='listview']")
			list.empty()
			$.each(result.news, function(index, oneNews) {
				var id = oneNews.id
				var title = oneNews.title
				var time = oneNews.updatedTime
				var newsHtml = "<li news-id='" + id + "'>" +
									"<a href='#'>" + 
										"<h3 style='white-space:normal;'>" + title + "</h3>" + 
										"<p>" + 	time + "</p>" + 
										"<span class='ui-li-count play-status' style='display:none;'>播报中...</span>" + 
									"</a>" + 
								"</li>"
	    			list.append(newsHtml)
	  		});
			list.listview('refresh');
		}
		
		hideLoading();
	});
	
	playStatusTimer = setInterval(playStatusTimerFunc,1000);
}

function playStatusTimerFunc() {
	$.getJSON("/news/getplayingnews", function(result) {
		$("span.play-status").hide()
		if (result.result == "success" && result.playStatus == "playing") {
			var newsId = result.newsId
			$("li[news-id='" + newsId + "'] span.play-status").show()
		}
	});
}

function showLoading(){
	$.mobile.loading('show', {
		text: '正在加载新闻...',
		textVisible: true,
		theme: 'a',
		html: ""
	});
}

function hideLoading() {
	$.mobile.loading('hide');
}
