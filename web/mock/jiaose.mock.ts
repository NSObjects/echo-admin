// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/roles': (req: Request, res: Response) => {
    res.status(200).send({ code: 64, msg: '情角共军下制存因叫政步党建。' });
  },
  'GET /api/roles': (req: Request, res: Response) => {
    res.status(200).send({
      code: 81,
      msg: '热子有那计证快加原江条三向。',
      data: {
        total: 63,
        list: [
          {
            name: '徐桂英',
            order: 75,
            identify: '价正劳存下无我十四品正图治区清自层。',
            state: 84,
          },
          { name: '萧刚', order: 99, identify: '学条律被领料九间边火九维领或统平。', state: 98 },
          { name: '邹军', order: 67, identify: '选式是长则或它适太影示育将年并领。', state: 93 },
          { name: '汤超', order: 76, identify: '面便为地真己龙清圆发目行车。', state: 68 },
          { name: '赵刚', order: 78, identify: '四特成千识劳形器导马得该问列。', state: 76 },
          { name: '胡艳', order: 74, identify: '整后别青品世县治又这红经白作热任西。', state: 98 },
          { name: '邱洋', order: 87, identify: '除导分联带无导组江身又代指物至。', state: 93 },
          { name: '秦勇', order: 68, identify: '何感社长天了七素成整革用前门队。', state: 93 },
          {
            name: '贺勇',
            order: 61,
            identify: '得但转高共听劳解满党至家建派例法引带。',
            state: 94,
          },
        ],
      },
    });
  },
  'DELETE /api/roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 67, msg: '正拉接维边人得下维边路小流音。' });
  },
  'PUT /api/roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 96, msg: '天级口联采但性老正规光济了团。' });
  },
  'PUT /api/roles/:id/menus': (req: Request, res: Response) => {
    res.status(200).send({ code: 76, msg: '日于证张们放目专即才铁众们温两照。' });
  },
};
