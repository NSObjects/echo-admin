// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /departments': (req: Request, res: Response) => {
    res.status(200).send({ code: 95, msg: '常极公也属象积成南表到题近过拉。' });
  },
  'GET /departments': (req: Request, res: Response) => {
    res.status(200).send({
      code: 86,
      msg: '和物月制积方问研生进及统同省整应。',
      data: {
        total: 71,
        list: [
          {
            name: '胡杰',
            parent_id: 78,
            email: 'k.kweb@csqn.fi',
            phone: '11283484151',
            status: 87,
            sort: 68,
            principal_id: 74,
          },
          {
            name: '易芳',
            parent_id: 80,
            email: 'j.rsvvqsj@cknootwl.mn',
            phone: '11223338171',
            status: 88,
            sort: 75,
            principal_id: 75,
          },
          {
            name: '孙强',
            parent_id: 60,
            email: 'e.rdvnvuyau@dlxdf.gu',
            phone: '11258822758',
            status: 95,
            sort: 85,
            principal_id: 95,
          },
          {
            name: '谭超',
            parent_id: 78,
            email: 'k.jfkhcgv@coslu.kg',
            phone: '11241383714',
            status: 73,
            sort: 68,
            principal_id: 98,
          },
        ],
      },
    });
  },
  'PUT /departments/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 61, msg: '听精照教较能一具规被器委育。' });
  },
  'GET /departments/:id': (req: Request, res: Response) => {
    res.status(200).send({
      code: 96,
      msg: '关干时干满程广装织江群律之总新果。',
      data: {
        name: '谭勇',
        parent_id: 84,
        email: 'w.dflod@gzousxur.au',
        phone: '11265897989',
        status: 87,
        sort: 93,
        principal_id: 93,
      },
    });
  },
  'DELETE /departments/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 72, msg: '山就军工受从半经们以门段集程社计术。' });
  },
};
