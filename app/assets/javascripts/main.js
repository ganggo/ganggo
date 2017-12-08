//= require javascripts/api
//= require javascripts/markdown
//= require javascripts/notification

// change border color if anchor is set
anchors = /#(.+?)$/.exec(window.location.href);
if (anchors !== null) {
  var postElem = $("a[name='" + anchors[1] + "']").closest(".panel-default");
  postElem.css("border-left-color", "#5bc0de");
  postElem.css("border-left-width", "5px");
}

// clicking the cross icon on any kind of
// alert should hide the element
$("#flash-container .alert").click(function() {
  $(this).hide();
});
$("#flash-container .alert-success").fadeOut(2000);
$("#flash-container .alert-danger").fadeOut(5000);

// find all like buttons and handle events
$(".comment-footer i").each(function(i, elem) {
  var postID = $(elem).attr("data-id");
  if (typeof postID === "undefined") {
    return;
  }
  API.posts(postID).likes.get().then(function(likes) {
    var likeCnt = 0;
    var dislikeCnt = 0;
    $.each(likes, function(i, like) {
      if (like.Positive) {
        likeCnt++;
      } else {
        dislikeCnt++;
      }
    });

    // set db count
    if ($(elem).hasClass("like")) {
      $(elem).html(likeCnt);
    } else {
      $(elem).html(dislikeCnt);
    }

    // register click event
    $(elem).click(function() {
      var positive = false;
      if ($(elem).hasClass("like")) {
        positive = true;
      }
      API.posts(postID).likes(positive).post().then(function () {
        var cnt = parseInt($(elem).text());
        $(elem).html(cnt+1);
      });
    });
  });
});
