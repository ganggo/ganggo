//= require javascripts/api
//= require javascripts/flash
//= require javascripts/markdown
//= require javascripts/notification

//= require javascripts/like_button
//= require javascripts/retweet_button
//= require javascripts/delete_button
//= require javascripts/userstreams

//= require javascripts/navigation

// change border color if anchor is set
anchors = /#(.+?)$/.exec(window.location.href);
if (anchors !== null) {
  var postElem = $("a[name='" + anchors[1] + "']").closest(".card");
  postElem.addClass("text-white bg-info");
  postElem.find(".card-header").addClass("text-white bg-info");
}
