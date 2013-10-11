var limit = 5
$(function() {
	$("#refresh").click(function(){
		showNews(true)
	});
	
	$("#play").click(function(){

	});

	$("#stop").click(function(){

	});

	$("#next").click(function(){
		// TODO send next request
	});
	showNews(false);
});

function showNews(refresh) {
	showLoading()
	var url = 'news/getnews?limit=' + limit
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
				var newsHtml = "<li news-id='" + id + "'><a href='#'>" + title + " [" + time + "]<span class='ui-li-count'>播放中...</span></a></li>"
	    			list.append(newsHtml)
	  		});
			list.listview('refresh');
		}
		
		hideLoading();
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
