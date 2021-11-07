(function(){
    var canvas=document.getElementById('canv');
    var cv=canvas.getContext('2d');
    cv.beginPath();
    cv.fillStyle = "#f5f5f5";
    cv.fillRect(0, 0, 700, 400);
    cv.lineJoin="round";
    var brush_size=3;
    var brush_color="#000000";
    var brush_alpha=1.0;

    var bef_x=""
    var bef_y=""

    function draw(x, y){
        x=~~x,y=~~y;
        cv.beginPath();
        if(bef_x==""){
            bef_x=x,bef_y=y;
            return;
        }
        if(bef_x==x && bef_y==y)return;
        cv.moveTo(bef_x,bef_y);
        cv.lineTo(x,y);
        cv.lineCap="round";
        cv.lineWidth=brush_size;
        cv.strokeStyle=brush_color;
        cv.stroke();
        bef_x=x,bef_y=y;
    }

    canvas.addEventListener('mousemove', e => {
        if(e.buttons&1){
            var rect=e.target.getBoundingClientRect();
            var x=e.clientX-rect.left;
            var y=e.clientY-rect.top;
            draw(x,y);
        }
    });

    canvas.addEventListener('mousedown', e => {
        if(e.button==0){
            var rect=e.target.getBoundingClientRect();
            var x=e.clientX-rect.left;
            var y=e.clientY-rect.top;
            draw(x,y);
        }
    });

    canvas.addEventListener('mouseup', drawEnd);
    canvas.addEventListener('mouseout', e => {
        if(e.buttons&1){
            drawEnd(e);
        }
    });

    function drawEnd(e){
        var rect=e.target.getBoundingClientRect();
        var x=e.clientX-rect.left;
        var y=e.clientY-rect.top;
        draw(x,y);
        bef_x="",bef_y="";
    }

}());