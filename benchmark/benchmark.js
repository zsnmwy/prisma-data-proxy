/*
  https://k6.io/docs/getting-started/installation/
  https://k6.io/docs/getting-started/running-k6/

  $ k6 run --vus 10 --duration 30s  benckmark.js
*/

import { check } from 'k6';
import http from 'k6/http';

export default function () {
    const url = 'http://graphql:4467';
    // const payload = `{"query":"mutation CreatePost {createOnePost(data: {title: \\"myPost\\" attr: '{"a":123}' author: {connect: {email: \\"jens@wundergraph.com\\"}}}){id title}}","operationName":"CreatePost"}`;
    // const payload = `{"query":"query AllPosts {findManyPost(take: 2500){id title createdAt}}","operationName":"CreatePost"}`;
    // const payload = `{"operationName":"testquery","variables":{},"query":"query testquery { findFirstPost {  id    content   attr    createdAt    updatedAt  }}"}`
    const payload = `{"query":"mutation {  createOneUser(data: {    email: \"2023-06-26T20:01:32.144Z123@email.com\"    posts: {      create: {        title: \"posts\"        attr: \"{\\\"a\\\":123,\\\"b\\\":true,\\\"c\\\":{\\\"d\\\":22,\\\"e\\\":\\\"123\\\"}}\"      }    }  }) {    id    email    name  }}","variables":{}}`

    const params = {
        headers: {
            'Content-Type': 'application/json',
            authorization: 'Bearer custometoken',
        },
    };

    const res = http.post(url, payload, params);
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    check(res, {
        'verify body': (r) =>
            r.body.includes('myPost'),
    });
}
