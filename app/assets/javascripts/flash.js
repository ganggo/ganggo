// display a status message
function notify(msg) {
  var div = $('<div class="alert alert-success mb-1">');
  div.text(msg);
  div.append('<i class="fa fa-times fa-lg pull-right">');

  $('#flash-container').append(div);
  div.fadeOut(5000);
}
