for (var ls = document.getElementsByTagName('A'), numLinks = ls.length, i=0; i<numLinks; i++){
	ls[i].setAttribute("onclick", "doalert(this); return false;")
}
function doalert(obj) {
	var sauce = obj.getAttribute("href");
	window.external.invoke(sauce);
	return false;
}
