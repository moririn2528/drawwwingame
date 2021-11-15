const domain_name="http://localhost:1213"

async function postData(url = "", data = {}) {
    console.log(domain_name+url);
    console.log(JSON.stringify(data));
    const response = await fetch(domain_name+url, {
        method: "POST",
        mode: "cors",
        //cache: "no-cache",
        headers: {
            //"Accept": "application/json",
            "Content-Type": "application/json;charset=UTF-8",
            "Origin": "file:///C:/Users/stran/Documents/GitHub/drawwwingame/src/html"
        },
        redirect: "follow",
        body: JSON.stringify(data),
    })
    return response;
}

async function createHashPassword(str){
    const uint8  = new TextEncoder().encode(str)
    const digest = await crypto.subtle.digest("SHA-256", uint8)
    return Array.from(new Uint8Array(digest)).map(v => v.toString(16).padStart(2,"0")).join("")
}