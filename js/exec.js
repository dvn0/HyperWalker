function addScript(uri) {
    return new Promise((resolve) => {
        const script = document.createElement("script");
        script.type= "text/javascript";
        script.src = uri;
        script.addEventListener("load", resolve) 
        document.body.appendChild(script);
    });
}
return (
    addScript("http://127.0.0.1:61628/js/freezedry/standalone.js").
        then ((html) => freezeDry.default()));
