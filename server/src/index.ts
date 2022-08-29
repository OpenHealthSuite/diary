import express from 'express';
import qs from 'qs';
import { FoodStorageRouter } from './handlers';
import { userMiddleware } from './middlewares';

const port = process.env.PORT || 3012;

const app = express();

app.settings['query parser'] = qs.parse

app.use(express.json())
app.use(express.static('./public'));

app.get('/api/health', (req, res) => {
  res.send()
});

app.use(userMiddleware);


app.use('/api', FoodStorageRouter)
  
const running = app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)

    process.on('SIGTERM', () => running.close());
    process.on('SIGINT', () => running.close()); 
    process.on('exit', () => running.close()); 
});