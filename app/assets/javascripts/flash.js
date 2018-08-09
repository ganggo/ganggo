// clicking the cross icon on any kind of
// alert should hide the element
$("#flash-container .alert").click(function() {
  $(this).hide();
});
$("#flash-container .alert i").unbind();
$("#flash-container .alert-success").fadeOut(2000);
// display a status message
function notify(msg) {
  var div = $('<div class="alert alert-success mb-1">');
  div.text(msg);
  div.append('<i class="fa fa-times fa-lg pull-right">');

  $('#flash-container').append(div);
  div.fadeOut(5000);
}
