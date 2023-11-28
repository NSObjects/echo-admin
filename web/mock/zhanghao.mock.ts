// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /login/account': (req: Request, res: Response) => {
    res.status(200).send({
      code: 60,
      msg: '万数合米会共做件周就收八观。',
      data: { token: '平在包条组团集却片这色才半却南。', type: 1 },
    });
  },
  'GET /api': (req: Request, res: Response) => {
    res.status(200).send([
      {
        method: '装公必院极了亲四场龙保整三。',
        path: '京革感级县电人团除次酸识候受装想法。',
        name: '冯艳',
      },
      {
        method: '解法治证酸战自族林矿报阶月处。',
        path: '公管型取空声厂了其些力置要。',
        name: '韩刚',
      },
      {
        method: '育完表听却强及千时积结法后心利。',
        path: '究周转总被性次和铁先养油业京专要。',
        name: '江娜',
      },
    ]);
  },
  'POST /login/out': (req: Request, res: Response) => {
    res.status(200).send({ code: 69, msg: '两采无省连联先最己查却放。' });
  },
};
