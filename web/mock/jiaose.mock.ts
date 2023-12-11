// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/roles': (req: Request, res: Response) => {
    res.status(200).send({ code: 88, msg: '着院少增好他劳给在们级机量当放。' });
  },
  'GET /api/roles': (req: Request, res: Response) => {
    res.status(200).send({
      code: 88,
      msg: '主不问单矿存构片林叫关气常约目。',
      data: {
        total: 84,
        list: [
          { name: '傅伟', order: 83, identify: '空观况部器八体土他打马进院。', state: 98 },
          { name: '罗磊', order: 81, identify: '少使候示党参定层土想县比支议才。', state: 92 },
          { name: '赖磊', order: 80, identify: '明利来即团研压史图回三特引场。', state: 71 },
          { name: '康刚', order: 67, identify: '作精它南快包区江拉两华给近有求龙识。', state: 85 },
          {
            name: '崔涛',
            order: 72,
            identify: '光求不劳社才反难小期极解统克时单土特。',
            state: 61,
          },
          { name: '张娜', order: 90, identify: '战争小质从济对则称真公价。', state: 70 },
          { name: '任秀英', order: 68, identify: '革它离部总需议形候土民图用山现走。', state: 61 },
          { name: '常勇', order: 77, identify: '问行表县政系业矿革委完教照。', state: 77 },
          { name: '雷洋', order: 64, identify: '节验工多公声百过根形身数百就高林。', state: 78 },
          { name: '余勇', order: 67, identify: '提何京儿然通些压活重同该安都教。', state: 85 },
          { name: '叶涛', order: 74, identify: '水战毛电原证适明与加角消线。', state: 87 },
          { name: '蔡艳', order: 97, identify: '规族资然个走北又国气世维。', state: 88 },
        ],
      },
    });
  },
  'DELETE /api/roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 80, msg: '群间代正际想风军门海四后见而。' });
  },
  'PUT /api/roles/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 99, msg: '行进带南专部象却知代取众满。' });
  },
  'PUT /api/roles/:id/menus': (req: Request, res: Response) => {
    res.status(200).send({ code: 78, msg: '离却适小节任府金技叫放且专。' });
  },
};
