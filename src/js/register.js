(function () {
    async function register_func(e){
        const username=document.getElementById("username").value;
        const password_row=document.getElementById("password").value;
        const password= await createHashPassword(password_row)
        const email=document.getElementById("email").value;
        document.getElementById("password").value="";
        console.log(username)
        console.log(password)
        postData("/register",{
            "username": username,
            "email": email,
            "password": password
        }).then(result=>{
            console.log("Data: ",result)
        }).catch(err=>{
            console.error("Error: ",err)
        })
    }

    register.addEventListener("click",register_func)
}());