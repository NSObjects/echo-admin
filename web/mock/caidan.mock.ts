// @ts-ignore
import { Request, Response } from 'express';

export default {
  'POST /menus': (req: Request, res: Response) => {
    res.status(200).send({ code: 61, msg: '根原引队报动来图织细主法标二。' });
  },
  'GET /menus': (req: Request, res: Response) => {
    res.status(200).send({
      code: 69,
      msg: '收如作员斯张然水列着平引生应类。',
      data: [
        {
          component: '住如生如厂设效子理需重今根划关度。',
          parent_id: 82,
          layout: true,
          path: '市适构毛水率下须律根但情器更音养。',
          name: '吴敏',
          redirect: '律即维然军光导据织美片生发由目放细。',
        },
        {
          component: '经很红约之电江收先精维进中先例所。',
          parent_id: 90,
          layout: true,
          path: '风识时性一史斯点切术委查标情。',
          name: '傅伟',
          redirect: '江强员建热志斗员红价别利按。',
        },
        {
          component: '查提清书写列展斯而九把任第无深。',
          parent_id: 80,
          layout: true,
          path: '流们任节路还情或太高开示几品代。',
          name: '黎敏',
          redirect: '更务革习东铁广阶例只新前持才特。',
        },
        {
          component: '效运更造与统选受去例百写置矿种为全真。',
          parent_id: 75,
          layout: false,
          path: '查响照过样除于西本却部油解直民。',
          name: '汤娟',
          redirect: '上军生走每能调其经解热原局集物。',
        },
        {
          component: '八建过律住发他至消称劳并识离志布院。',
          parent_id: 96,
          layout: false,
          path: '想造压也周场常具资团是况将习六月决。',
          name: '梁强',
          redirect: '青层阶建半心程利存长情儿量前自局青。',
        },
      ],
    });
  },
  'PUT /menus/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 73, msg: '政位条采只立铁理八标据到始积但事张。' });
  },
  'DELETE /menus/:id': (req: Request, res: Response) => {
    res.status(200).send({ code: 86, msg: '子十易活来比所外持应真交新报。' });
  },
};
