// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /api/login/account': (req: Request, res: Response) => {
    res.status(200).send({
      code: 64,
      msg: '道响委少真第最间做会农数前主本历。',
      data: { token: '这离五变亲据过称过离厂区又观无飞。', type: 1 },
    });
  },
  'GET /api/api': (req: Request, res: Response) => {
    res.status(200).send([
      {
        method: '如音看派向世列程容还共商极。',
        path: '还至切天问老片备提电儿济备极却。',
        name: '邓娟',
      },
      {
        method: '直才且证广了华标县心用并速天响术。',
        path: '例切商过到被性积路织情各族听儿。',
        name: '卢强',
      },
      {
        method: '教且那决直速众声金目美石分同。',
        path: '基主研线社月压连设回政长任。',
        name: '汤军',
      },
      {
        method: '约状亲空形是再派明业指为二造果性反。',
        path: '价值速物约置造反标反员内义己现之但山。',
        name: '潘娟',
      },
      {
        method: '名生战太起低往指八系边电思家值。',
        path: '能上标习还包能你适同亲容说对斗所。',
        name: '马芳',
      },
      {
        method: '状族前火交水长石清适广山使严元没党。',
        path: '参加会器思王率酸什按联电头任值路器。',
        name: '冯霞',
      },
      {
        method: '程往基西却织产转规先长总领。',
        path: '律民书什族结带动按情小以员方片。',
        name: '朱勇',
      },
      {
        method: '表马两任生难再照小求县光平需南。',
        path: '他照品重期铁对县易算状周况管风总风非。',
        name: '阎涛',
      },
      {
        method: '往金性十长克维角又时型众。',
        path: '备作义以体结着声学增安制系议产。',
        name: '姜娜',
      },
      {
        method: '生里照务好指展半果所律因来色知级做。',
        path: '手间此数整器率律家活斗新它。',
        name: '吴敏',
      },
      {
        method: '县开亲院运型构细青转团类边教。',
        path: '江究量线识两儿车专流带各证平。',
        name: '汤秀兰',
      },
      {
        method: '科步几叫品周存然全类总记老。',
        path: '从把重标很它农身老节今多么。',
        name: '石刚',
      },
      {
        method: '影照越离回业查是加准族非历劳。',
        path: '调动府加又四此京花完真步样。',
        name: '蔡秀英',
      },
      {
        method: '积数线包受使流义员论听种北三写。',
        path: '产气于南于例该速半走术话属山我区满。',
        name: '赖艳',
      },
      {
        method: '此要国于属四全派思权合转题并制示想专。',
        path: '立万位两成到劳本社南小持国风才响。',
        name: '夏刚',
      },
      {
        method: '决具今离参造公等类且走酸过矿包。',
        path: '向较须将具毛员好反达值共现当住立。',
        name: '常杰',
      },
      {
        method: '工小治此科定点地被来布众思三。',
        path: '年响一过部运大万节资院月题。',
        name: '赵磊',
      },
      {
        method: '十那候六平子处律团常都立间起中得又。',
        path: '正反构处眼济连业实各张而花带任。',
        name: '丁超',
      },
    ]);
  },
  'POST /api/login/out': (req: Request, res: Response) => {
    res.status(200).send({ code: 93, msg: '具社采解放物节计率队不经。' });
  },
};
