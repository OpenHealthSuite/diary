import express from 'express';
import qs from 'qs';
import { FoodStorageRouter } from './handlers';
import { userMiddleware } from './middlewares';
import { closeCassandra, doCassandraMigrations, setDefaultGlobalClient } from './storage/cassandra';

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

setDefaultGlobalClient();

doCassandraMigrations().then(() => {
  const running = app.listen(port, () => {
    console.log(`Example app listening on port ${port}`)

    const cleanup = () => {
      closeCassandra()
      running.close()
    }

    process.on('SIGTERM', cleanup);
    process.on('SIGINT', cleanup); 
    process.on('exit', cleanup); 
  });
})
  