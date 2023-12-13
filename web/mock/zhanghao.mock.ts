// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/login/account': (req: Request, res: Response) => {
    res.status(200).send({
      code: 100,
      msg: '天应明面全思治何做确目两有。',
      data: { token: '布种之指队书人事应选大联亲长。', type: 1 },
    });
  },
  'GET /api/api': (req: Request, res: Response) => {
    res.status(200).send([
      {
        method: '原深至处查集容人交写度收九进速每。',
        path: '不们员调但活维斯品务那参团选美活酸。',
        name: '范娟',
      },
      {
        method: '革了革米许算价元红府个离。',
        path: '除深标月你次二除学此色技无变九。',
        name: '唐强',
      },
      {
        method: '重命认热就百备部增通连力次边非接五。',
        path: '员四要矿革整效决争置族率想细分。',
        name: '薛丽',
      },
      {
        method: '件具价品第铁织般里万少着应多人。',
        path: '定强节效历段放道重具响会群选打。',
        name: '陈秀英',
      },
      {
        method: '议标府又界志三东先热义引个传证划。',
        path: '习方却历今领地意族量气研参边必。',
        name: '黎明',
      },
      {
        method: '别龙类全每温酸儿确拉标易适好派有好只。',
        path: '物放如一级准情常感称引维清还角象并。',
        name: '孔静',
      },
      {
        method: '公断由地组业面亲建完接带内先越名再。',
        path: '受没团感电问直手周持争会上制且白非。',
        name: '胡伟',
      },
      {
        method: '始次理原百入间专该命计之必。',
        path: '她信后员十但放龙全了质着权书权自任理。',
        name: '朱芳',
      },
      {
        method: '建四引会实据置全与片百压现数十。',
        path: '形保然何理表但值样期合外电认。',
        name: '文超',
      },
      {
        method: '作志社线带名已查所十时将始包九毛程。',
        path: '铁报示段证状基真要向料日统关任。',
        name: '马杰',
      },
      {
        method: '入划江热工成千与面领五半海。',
        path: '必认连低开十段在取系步史亲相号量。',
        name: '吴强',
      },
      {
        method: '难求育须张者是例京青们此属结况。',
        path: '与提有业放适由公委个领几干花实热万科。',
        name: '沈明',
      },
    ]);
  },
  'POST /api/login/out': (req: Request, res: Response) => {
    res.status(200).send({ code: 71, msg: '始是声期养号她写利回图进样放。' });
  },
};
