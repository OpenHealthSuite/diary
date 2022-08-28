import express from 'express';
import qs from 'qs';
import { FoodStorageRouter } from './handlers';
import { userMiddleware } from './middlewares';

const port = process.env.PORT || 3012;

const app = express();

app.settings('query parser', function (str: string) {
  return qs.parse(str, { })
})

app.use(express.json())
app.use(express.static('./public'));
app.use(userMiddleware);

app.get('/api', (req, res) => {
    res.send('Hello World!')
});

app.use('/api', FoodStorageRouter)
  
const running = app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)

    process.on('SIGTERM', () => running.close());
    process.on('SIGINT', () => running.close()); 
    process.on('exit', () => running.close()); 
});