import express from 'express';
import { userMiddleware } from './middlewares';

const port = process.env.PORT || 3012;

const app = express();

app.use(express.json())
app.use(express.static('./public'));
app.use(userMiddleware);

app.get('/api', (req, res) => {
    res.send('Hello World!')
});

  
const running = app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)

    process.on('SIGTERM', () => running.close());
    process.on('SIGINT', () => running.close()); 
    process.on('exit', () => running.close()); 
});