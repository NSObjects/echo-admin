// @ts-ignore
import { Request, Response } from 'express';

export default {
  'GET /api/users': (req: Request, res: Response) => {
    res.status(200).send({
      code: 60,
      msg: '革也法满没构际农商更即增方路。',
      data: {
        total: 81,
        list: [
          {
            name: '杜芳',
            phone: '11237317235',
            status: 94,
            password: 'string(16)',
            account: '住个作老况局火众压部商力现。',
            avatar: 'https://avatars1.githubusercontent.com/u/8186664?s=40&v=4',
            role_id: 86,
            department_id: 80,
          },
          {
            name: '孔敏',
            phone: '11256241624',
            status: 71,
            password: 'string(16)',
            account: '农道文引厂去进产今离发何论。',
            avatar: '',
            role_id: 95,
            department_id: 75,
          },
          {
            name: '魏涛',
            phone: '11145756960',
            status: 74,
            password: 'string(16)',
            account: '意争议与段住原权话自加精度其交速。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
            role_id: 83,
            department_id: 73,
          },
          {
            name: '陆秀兰',
            phone: '11282884857',
            status: 78,
            password: 'string(16)',
            account: '京条已指利中老知较民历团我业。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/udxAbMEhpwthVVcjLXik.png',
            role_id: 86,
            department_id: 63,
          },
          {
            name: '锺丽',
            phone: '11479581493',
            status: 97,
            password: 'string(16)',
            account: '人义采严往最越教联青照易斯石。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png',
            role_id: 82,
            department_id: 74,
          },
          {
            name: '蔡磊',
            phone: '11267084423',
            status: 88,
            password: 'string(16)',
            account: '家去示况从得去话支已即处社术。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
            role_id: 76,
            department_id: 83,
          },
          {
            name: '董静',
            phone: '11204833626',
            status: 79,
            password: 'string(16)',
            account: '变直行但将离片查系各么名技文习性没。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
            role_id: 72,
            department_id: 90,
          },
          {
            name: '康敏',
            phone: '11181267320',
            status: 70,
            password: 'string(16)',
            account: '传再精半争太较民算式问响越置要。',
            avatar: '',
            role_id: 68,
            department_id: 62,
          },
          {
            name: '韩刚',
            phone: '11212083942',
            status: 94,
            password: 'string(16)',
            account: '再由老非式是属十养深率设里须。',
            avatar: 'https://avatars0.githubusercontent.com/u/507615?s=40&v=4',
            role_id: 61,
            department_id: 77,
          },
          {
            name: '赖娟',
            phone: '11216813017',
            status: 68,
            password: 'string(16)',
            account: '照国很点间新好经部步如青斯因。',
            avatar:
              'https://gw.alipayobjects.com/zos/antfincdn/XAosXuNZyF/BiazfanxmamNRoxxVxka.png',
            role_id: 67,
            department_id: 86,
          },
          {
            name: '姚刚',
            phone: '11151818132',
            status: 89,
            password: 'string(16)',
            account: '备算低议设越交级铁回包布拉中决。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
            role_id: 74,
            department_id: 88,
          },
          {
            name: '侯娟',
            phone: '11288426638',
            status: 63,
            password: 'string(16)',
            account: '效边细效象火开再所速特度片进难手身。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/udxAbMEhpwthVVcjLXik.png',
            role_id: 83,
            department_id: 71,
          },
          {
            name: '韩静',
            phone: '11243366528',
            status: 95,
            password: 'string(16)',
            account: '老影容百机离压省比表级你每保际为。',
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/ThXAXghbEsBCCSDihZxY.png',
            role_id: 93,
            department_id: 81,
          },
        ],
      },
    });
  },
  'POST /api/users': (req: Request, res: Response) => {
    res.status(200).send({ code: 89, msg: '十主用小想们始极气转东便真织节必形。' });
  },
  'GET /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({
      code: 68,
      msg: '用式表四示类这使断热空达经她光见风作。',
      data: {
        name: '阎平',
        phone: '11441607551',
        status: 78,
        password: 'string(16)',
        account: '本局了并计实用内力没地可话。',
        avatar: 'https://gw.alipayobjects.com/zos/rmsportal/udxAbMEhpwthVVcjLXik.png',
        role_id: 67,
        department_id: 78,
      },
    });
  },
  'PUT /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 92, msg: '证带或红离积听展细么传统展边除根于。' });
  },
  'DELETE /api/users/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 63, msg: '院经己族精当布图高步题划道角发知队统。' });
  },
  'GET /api/users/current': (req: Request, res: Response) => {
    res.status(200).send({
      code: 77,
      msg: '前争展标消华人江开声议相音立达。',
      data: {
        name: '尹秀兰',
        phone: '11411651212',
        status: 75,
        password: 'string(16)',
        account: '热路变根入海农平也北见得细管。',
        avatar: 'https://gw.alipayobjects.com/zos/rmsportal/OKJXDXrmkNshAMvwtvhu.png',
        role_id: 76,
        department_id: 77,
      },
    });
  },
};
