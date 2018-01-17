// on confirm trigger reshare api call
$("i.fa-times").click(function() {
  if (!confirm(msg("javascript.delete_confirm"))) {
    return;
  }

  var elem = $(this);
  var postID = $(this).attr("data-postID");
  var commentID = $(this).attr("data-commentID");
  if (typeof postID !== "undefined") {
    // NOTE delete workaround cause of:
    // https://github.com/yui/yuicompressor/issues/47
    var promise = API.posts(postID)['delete'];
    promise().then(function() {
      notify(msg("javascript.message_deleted"));
      elem.remove();
    });
  } else if (typeof commentID !== "undefined") {
    var promise = API.comments(commentID)['delete'];
    promise().then(function() {
      notify(msg("javascript.message_deleted"));
      elem.remove();
    });
  }

  return false;
});
