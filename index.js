const app = require("express")();
const server = require("http").createServer(app);
const cors = require("cors");

const io = requrie("socket.io")(server, {
    cors: {
        origin: "*",
        methoda: ["GET", "POST"]
    }
});


app.use(cors());

const PORT = process.env.PORT || 3000;






