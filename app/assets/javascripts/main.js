//= require javascripts/api
//= require javascripts/flash
//= require javascripts/markdown
//= require javascripts/notification

//= require javascripts/like_button
//= require javascripts/retweet_button
//= require javascripts/delete_button

// change border color if anchor is set
anchors = /#(.+?)$/.exec(window.location.href);
if (anchors !== null) {
  var postElem = $("a[name='" + anchors[1] + "']").closest(".card");
  postElem.addClass("text-white bg-info");
  postElem.find(".card-header").addClass("text-white bg-info");
}

// clicking the cross icon on any kind of
// alert should hide the element
$("#flash-container .alert").click(function() {
  $(this).hide();
});
$("#flash-container .alert-success").fadeOut(2000);
$("#flash-container .alert-danger").fadeOut(5000);
