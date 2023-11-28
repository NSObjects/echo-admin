// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /roles': (req: Request, res: Response) => {
    res.status(200).send({ code: 88, msg: '已提油五府况器规调候立白色应。' });
  },
  'GET /roles': (req: Request, res: Response) => {
    res.status(200).send({
      code: 92,
      msg: '部教观风务半别美院界何公写安里海温史。',
      data: {
        total: 70,
        list: [
          {
            name: '李明',
            order: '展一安产易根名发今专除意克际教上难。',
            identify: '本细器因产深花段包置之派保。',
            state: 84,
          },
          {
            name: '魏静',
            order: '其到科时增今看列什达响或工组所。',
            identify: '完更决红可究图图是决家安步线入百加。',
            state: 74,
          },
          {
            name: '蔡平',
            order: '传建代持第米联应元没除商或信维备。',
            identify: '个种民深需不商相众给维候给。',
            state: 78,
          },
          {
            name: '乔伟',
            order: '得引除总都义多务数史史程。',
            identify: '能了必机习满矿话布例类门规。',
            state: 68,
          },
          {
            name: '卢丽',
            order: '展除组说究心社除教也正据话。',
            identify: '素亲律高例新即群难党指容决院体。',
            state: 62,
          },
        ],
      },
    });
  },
  'DELETE /roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 96, msg: '员究压共精出只称影低厂话书了样。' });
  },
  'PUT /roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 81, msg: '证地速没效志还型共再流革。' });
  },
  'PUT /roles/:id/menus': (req: Request, res: Response) => {
    res.status(200).send({ code: 97, msg: '调能传成生眼头办却又形西者过院面准。' });
  },
};
