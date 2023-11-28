// @ts-ignore
import { Request, Response } from 'express';

export default {
  'GET /users': (req: Request, res: Response) => {
    res.status(200).send({
      code: 97,
      msg: '用有等发复强从查之传连平起国你化复细。',
      data: {
        total: 97,
        list: [
          {
            name: '崔静',
            phone: '11298697376',
            status: 81,
            password: 'string(16)',
            account: '家连被作矿没任难务行么段周里切规型。',
            avatar: 'https://avatars1.githubusercontent.com/u/8186664?s=40&v=4',
            id: '338fFA0D-D9be-d324-d59d-a18AA94f131b',
          },
          {
            name: '康秀英',
            phone: '11417961265',
            status: 60,
            password: 'string(16)',
            account: '常看们什毛有结特水安义派达定路做米。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
            id: '00F7f769-5eb3-613d-DDD1-2377d25e8A9D',
          },
          {
            name: '周强',
            phone: '11104566507',
            status: 78,
            password: 'string(16)',
            account: '原县无很料打细飞二路场圆用见布自来土。',
            avatar: 'https://avatars0.githubusercontent.com/u/507615?s=40&v=4',
            id: '9524b7FD-ec3f-1C89-2a73-8ecD5E7e1B9c',
          },
        ],
      },
    });
  },
  'POST /users': (req: Request, res: Response) => {
    res.status(200).send({ code: 87, msg: '好料要步水入率界看马元等路权三论。' });
  },
  'GET /users/:id': (req: Request, res: Response) => {
    res.status(200).send({
      code: 80,
      msg: '思才本满到改门办商形进军好生矿带共律。',
      data: {
        name: '乔芳',
        phone: '11287862623',
        status: 92,
        password: 'string(16)',
        account: '系引标定油油及十发效强月信那打。',
        avatar: 'https://avatars0.githubusercontent.com/u/507615?s=40&v=4',
        id: 'bC89cEd4-bDd0-7eCA-9d49-36D9c3E1Fd0F',
      },
    });
  },
  'PUT /users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 78, msg: '却市往队入土资型团发很员区住能级。' });
  },
  'DELETE /users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 60, msg: '解取力效习定白体们思设地队况毛。' });
  },
  'GET /users/current': (req: Request, res: Response) => {
    res.status(200).send({
      code: 100,
      msg: '儿克点治大细花以般军周明。',
      data: {
        name: '王涛',
        phone: '11167886641',
        status: 69,
        password: 'string(16)',
        account: '气现得离于管决看老因说使中。',
        avatar: 'https://avatars1.githubusercontent.com/u/8186664?s=40&v=4',
        id: 'cdc7A6Ae-9a6B-aeB4-bC4D-c7b6cE8Ab0dB',
      },
    });
  },
};
