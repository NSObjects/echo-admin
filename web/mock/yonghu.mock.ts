// @ts-ignore
import { Request, Response } from 'express';

export default {
  'GET /api/users': (req: Request, res: Response) => {
    res.status(200).send({
      code: 97,
      msg: '件两片理然市数养太先整真率达它低。',
      data: {
        total: 68,
        list: [
          {
            name: '袁洋',
            phone: '11445535602',
            status: 74,
            password: 'string(16)',
            account: '常治法置容统就九习争联更别去会。',
            avatar: 'https://avatars0.githubusercontent.com/u/507615?s=40&v=4',
            role_id: 79,
            department_id: 71,
          },
          {
            name: '蒋明',
            phone: '11263853255',
            status: 80,
            password: 'string(16)',
            account: '身意且却活标龙属持接部主月低政花劳。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png',
            role_id: 70,
            department_id: 68,
          },
          {
            name: '常霞',
            phone: '11281418925',
            status: 87,
            password: 'string(16)',
            account: '好那约关观酸工过国反员儿产。',
            avatar:
              'https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png',
            role_id: 90,
            department_id: 91,
          },
          {
            name: '范刚',
            phone: '11283110867',
            status: 96,
            password: 'string(16)',
            account: '样素成速布争统音情真多阶给己导叫率增。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/KDpgvguMpGfqaHPjicRK.svg',
            role_id: 66,
            department_id: 76,
          },
          {
            name: '郭静',
            phone: '11241612645',
            status: 78,
            password: 'string(16)',
            account: '分院且的清儿众做造育眼院速住。',
            avatar: 'https://avatars1.githubusercontent.com/u/8186664?s=40&v=4',
            role_id: 73,
            department_id: 93,
          },
        ],
      },
    });
  },
  'POST /api/users': (req: Request, res: Response) => {
    res.status(200).send({ code: 80, msg: '证八说始按军连县次形动方。' });
  },
  'GET /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({
      code: 90,
      msg: '干全外值离过则活养科你并历海音。',
      data: {
        name: '高勇',
        phone: '11260786865',
        status: 65,
        password: 'string(16)',
        account: '部音色她价报识达热部向一。',
        avatar: 'https://gw.alipayobjects.com/zos/rmsportal/udxAbMEhpwthVVcjLXik.png',
        role_id: 100,
        department_id: 61,
      },
    });
  },
  'PUT /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 88, msg: '写行思先包力自别较严斗消。' });
  },
  'DELETE /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 64, msg: '红装自红从会导公与严因群期里程。' });
  },
  'GET /api/users/current': (req: Request, res: Response) => {
    res.status(200).send({
      code: 98,
      msg: '特系接没北斯代养提亲事离习我。',
      data: {
        name: '张超',
        phone: '11186816299',
        status: 79,
        password: 'string(16)',
        account: '变业更近等光路状非难只律识后采着身光。',
        avatar: 'https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png',
        role_id: 64,
        department_id: 76,
      },
    });
  },
};
