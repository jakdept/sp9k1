function close() {
  // $(".preview-container").hide();
  document.getElementById("preview-container").style.display = "none";
  $(".image-container").removeClass("blur");
}

function open(target) {
  console.log("opening an image");
  document.getElementById("preview-pane").src = target.getAttribute("data-original");
  document.getElementById("preview-caption").innerHTML = target.alt;
  document.getElementById("preview-container").style.display = "flex";
  $(".image-container").addClass("blur");
  // $("#preview-container").show(0);
}

$(document).ready(function () {

  document.getElementById("preview-container").onclick = close;
  // imageCards = document.getElementsByClassName("image-card");
  // for (var i = 0; i < imageCards.length; i++) {
  //   imageCards[i].onflick = function() {
  //     console.log("clicked an image")
  //     open(this);
  //   };
  // }

  $(".image-card").on("click", function () {
    console.log("clicked an image");
    open(this);
  });

  // $("#preview-pane").hide(0);
  // $("#preview-container").hide(0);

  $(window).resize(close)
  $(document).keyup(function (e) {
    if (e.keyCode == 27) {
      // keycode 27 is escape
      close();
    }
  });
});