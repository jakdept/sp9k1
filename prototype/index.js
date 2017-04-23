function close() {
  // $(".preview-pane").hide(10);
  $("#preview-pane-container").hide(100, function() {
  // $("#preview-pane-container").slideUp(200, function() {
    $(".image-card-selected").animate({ marginBottom: "10px" }, 200);
  });
}

$(document).ready(function () {

  $(".preview-pane-close-button ").on("click", close); 

});