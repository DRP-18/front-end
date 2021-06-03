const app = require("express")();
const server = require("http").createServer(app);
const cors = require("cors");

const io = require("socket.io")(server, {
    cors: {
        origin: "*",
        methoda: ["GET", "POST"]
    }
});


app.use(cors());

const PORT = process.env.PORT || 5000;

app.get("/", (req, res) => {
    res.send(`Server is running on port ${PORT}`);
});


io.on("connection", socket => {
    /* Will emit socket ID as soon as the connection opens*/
    socket.emit("me", socket.id);

    socket.on("disconnect", () => {
        socket.broadcast.emit("callended");
    });

    socket.on("calluser", ({userToCall, signalData, from, name}) => {
        io.to(userToCall).emit("calluser", {signal:signalData, from, name});
    });

    socket.on("answercall", (data) => {
        io.to(data.to).emit("callaccepted", data.signal);
    });
});


server.listen(PORT, "0.0.0.0", () => console.log(`Server is listening on port ${PORT}`));





