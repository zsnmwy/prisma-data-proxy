import { PrismaClient } from '@prisma/client/edge'

const db = new PrismaClient({
    datasources: {
        db: {
            url: "prisma://https-portal4466/?api_key=custometoken"
        }
    }
});

(async () => {
    await db.user.create({
        data: {
            email:  new Date().toISOString()+"123@email.com",
            posts: {
                create: {
                    title: "posts",
                    attr: {
                        a: 123,
                        b: true,
                        c: {
                            d: 22,
                            e: "123"
                        }
                    }
                }
            }
        }
    })

    const res = await db.post.findMany()
    console.log(res)

    const userRes = await db.user.findMany()
    console.log(userRes)
})()