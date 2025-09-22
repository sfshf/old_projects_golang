import express from 'express'
import cors from 'cors'
import request from 'sync-request'
import http2 from 'http2'
import fs from 'fs'
import nodemailer from 'nodemailer'
import moment from 'moment'

const app = express()
const port = 4010

// app.use(cors())

const transporter = nodemailer.createTransport({
  host: 'box.n1xt.net', // SMTP服务器地址
  port: 465, // SMTP服务器端口号
  secureConnection: true, // 不使用SSL加密连接（如果需要，设置为true）
  auth: {
    user: 'noreply@n1xt.net', // 发件人邮箱地址
    pass: 'xernyh-hyktyg13' // 发件人邮箱密码或者应用程序特定密码
  }
})

const emailCount = {}

const emailList = ["gavin@n1xt.net", "sheldon@n1xt.net"]

const sendSuccessMail = (endpoint) => {
  for (let i = 0; i < emailList.length; i++) {
    transporter.sendMail({
      from: 'noreply@n1xt.net', // 发件人名称及邮箱地址
      to: emailList[i], // 收件人邮箱地址
      subject: 'Status Success', // 邮件主题
      text: endpoint + ' endpoint check success' // 邮件正文
    }, (error, info) => {
      if (error) {
        console.log(`Error occurred while sending the email: ${error}`)
      } else {
        console.log(`Email sent successfully! Message ID: ${info.messageId}`)
      }
    })
  }
}

const sendFailMail = (endpoint) => {
  for (let i = 0; i < emailList.length; i++) {
    transporter.sendMail({
      from: 'noreply@n1xt.net', // 发件人名称及邮箱地址
      to: emailList[i], // 收件人邮箱地址
      subject: 'Status Error', // 邮件主题
      text: endpoint + ' endpoint check fail' // 邮件正文
    }, (error, info) => {
      if (error) {
        console.log(`Error occurred while sending the email: ${error}`)
      } else {
        console.log(`Email sent successfully! Message ID: ${info.messageId}`)
      }
    })
  }
}

const handleHttp1Request = (
  url,
  method,
  reqData,
  respHandle,
  respDataHandle
) => {
  let opts = {}
  if (reqData) {
    opts = {
      json: reqData
    }
  }
  var res = request(method, url, opts)
  if (respHandle) {
    respHandle(res)
  }
  if (respDataHandle) {
    respDataHandle(res.getBody('utf8'))
  }
}

const statusHttpRespCode = (status, idx, address, method, param, respCode) => {
  let currentDT = moment().unix()
  status.endpoints[idx].currentDT = currentDT
  try {
    handleHttp1Request(
      address,
      method,
      param,
      (resp) => {
        if (resp.statusCode === respCode) {
          if (status.endpoints[idx].status == false) {
            status.endpoints[idx].status = true
            status.endpoints[idx].lastDT = currentDT
          }
          status.proportion += 1;
          if (emailCount[address] === 1) {
            sendSuccessMail(address)
            emailCount[address] = 0
          }
        } else {
          if (status.endpoints[idx].status == true) {
            status.endpoints[idx].status = false
            status.endpoints[idx].lastDT = currentDT
          }
          if (emailCount[address] === 0) {
            sendFailMail(address)
            emailCount[address] = 1
          }
        }
      },
      null,
    )
  } catch (err) {
    if (status.endpoints[idx].status == true) {
      status.endpoints[idx].status = false
      status.endpoints[idx].lastDT = currentDT
    } else if (!status.endpoints[idx].lastDT) {
      status.endpoints[idx].lastDT = currentDT
    }
    if (emailCount[address] === 0) {
      sendFailMail(address)
      emailCount[address] = 1
    }
  }
}

const handleHttp2Request = (
  hostname,
  path,
  method,
  reqData,
  respHandle,
  respDataHandle
) => {
  const reqOpts = {
    ':path': path,
    ':method': method
  }
  let postData = null
  if (reqData) {
    postData = JSON.stringify(reqData)
    reqOpts['content-type'] = 'application/json'
    reqOpts['content-length'] =  Buffer.byteLength(postData)
  }
  const client = http2.connect(hostname, {
    ca: fs.readFileSync('./certs.pem'),
  });
  const req = client.request(reqOpts)
  req.setEncoding('utf8');
  if (respHandle) {
    req.on('response', respHandle)
  }
  if (respDataHandle) {
    req.on('data', respDataHandle)
  }
  req.on('error', (e) => {
    console.log("http2 request error:", e)
    client.close()
  })
  if (postData) {
    req.write(postData)
  }
  req.on('end', () => {
    client.close()
  })
  req.end()
}

const statusHttp2RespCode = (status, idx, hostname, path, method, param, respCode) => {
  let address = hostname + path
  let currentDT = moment().unix()
  status.endpoints[idx].currentDT = currentDT
  try {
    handleHttp2Request(
      hostname,
      path,
      method,
      param,
      (headers, flags) => {
        if (headers[':status'] == respCode) {
          if (status.endpoints[idx].status == false) {
            status.endpoints[idx].status = true
            status.endpoints[idx].lastDT = currentDT
          }
          status.proportion += 1;
          if (emailCount[address] === 1) {
            sendSuccessMail(address)
            emailCount[address] = 0
          }
        } else {
          if (status.endpoints[idx].status == true) {
            status.endpoints[idx].status = false
            status.endpoints[idx].lastDT = currentDT
          }
          if (emailCount[address] === 0) {
            sendFailMail(address)
            emailCount[address] = 1
          }
        }
      },
      null,
    )
  } catch (err) {
    if (status.endpoints[idx].status == true) {
      status.endpoints[idx].status = false
      status.endpoints[idx].lastDT = currentDT
    } else if (!status.endpoints[idx].lastDT) {
      status.endpoints[idx].lastDT = currentDT
    }
    if (emailCount[address] === 0) {
      sendFailMail(address)
      emailCount[address] = 1
    }
  }
}

const statusHttp2RespDataCode = (status, idx, hostname, path, method, param, respDataCode) => {
  let address = hostname + path;
  let currentDT = moment().unix()
  status.endpoints[idx].currentDT = currentDT
  try {
    handleHttp2Request(
      hostname,
      path,
      method,
      param,
      null,
      (chunk) => {
        const respData = JSON.parse(chunk)
        if (respData.code === respDataCode) {
          if (status.endpoints[idx].status == false) {
            status.endpoints[idx].status = true
            status.endpoints[idx].lastDT = currentDT
          }
          status.proportion += 1;
          if (emailCount[address] === 1) {
            sendSuccessMail(address)
            emailCount[address] = 0
          }
        } else {
          if (status.endpoints[idx].status == true) {
            status.endpoints[idx].status = false
            status.endpoints[idx].lastDT = currentDT
          }
          if (emailCount[address] === 0) {
            sendFailMail(address)
            emailCount[address] = 1
          }
        }
      },
    )
  } catch (err) {
    if (status.endpoints[idx].status == true) {
      status.endpoints[idx].status = false
      status.endpoints[idx].lastDT = currentDT
    } else if (!status.endpoints[idx].lastDT) {
      status.endpoints[idx].lastDT = currentDT
    }
    if (emailCount[address] === 0) {
      sendFailMail(address)
      emailCount[address] = 1
    }
  }
}

app.post("/status/all", (req, res) => {
  const allStatus = [
    kongStatus,
    discourseStatus,
    jenkinsStatus,
    slarkUIStatus,
    slarkStatus,
    wordStatus,
  ]
  const data = {
    now: moment().unix(),
    list: allStatus
  }
  res.status(200).send(data)
})

const kongAddress = 'https://api.n1xt.net'
const discourseAddress = 'http://vplus.forum.n1xt.net:8080'
const jenkinsAddress = 'http://jenkins.n1xt.net:8500'
const slarkUIAddress = 'http://sso.n1xt.net'

// kong endpoints -------------------------------------------------------------------
const kongStatus = {
  name: "kong",
  proportion: 0,
  endpoints: [
    {
      name: "Kong",
      hostname: kongAddress,
      path: "/",
      method: "GET",
      param: null,
      respCode: '404',
      status: false,
      lastDT: 0,
      currentDT: 0
    }
  ]
}

// discourse endpoints -------------------------------------------------------------------
const discourseStatus = {
  name: "discourse",
  proportion: 0,
  endpoints: [
    {
      name: "Discourse",
      hostname: discourseAddress,
      path: "/",
      method: "GET",
      param: null,
      respCode: 200,
      status: false,
      lastDT: 0,
      currentDT: 0
    }
  ]
}

// jenkins endpoints -------------------------------------------------------------------
const jenkinsStatus = {
  name: "jenkins",
  proportion: 0,
  endpoints: [
    {
      name: "Jenkins",
      hostname: jenkinsAddress,
      path: "/",
      method: "GET",
      param: null,
      respCode: 403,
      status: false,
      lastDT: 0,
      currentDT: 0
    }
  ]
}

// slark ui endpoints -------------------------------------------------------------------
const slarkUIStatus = {
  name: "slarkUI",
  proportion: 0,
  endpoints: [
    {
      name: "SlarkUI",
      hostname: slarkUIAddress,
      path: "/",
      method: "GET",
      param: null,
      respCode: 200,
      status: false,
      lastDT: 0,
      currentDT: 0
    }
  ]
}

// slark endpoints -------------------------------------------------------------------
const slarkStatus = {
  name: "slark",
  proportion: 0,
  endpoints: [
    {
      name: "LoginByPhone",
      hostname: kongAddress,
      path: "/slark/user/loginByPhone/v1",
      method: "POST",
      param: null,
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "LoginByEmail",
      hostname: kongAddress,
      path: "/slark/user/loginByEmail/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "LoginBySession",
      hostname: kongAddress,
      path: "/slark/user/loginBySession/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "LoginByApple",
      hostname: kongAddress,
      path: "/slark/user/loginByApple/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "RegisterByEmail",
      hostname: kongAddress,
      path: "/slark/user/registerByEmail/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "SendRegistrationEmailCaptcha",
      hostname: kongAddress,
      path: "/slark/user/sendRegistrationEmailCaptcha/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "Logout",
      hostname: kongAddress,
      path: "/slark/user/logout/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "CheckLogin",
      hostname: kongAddress,
      path: "/slark/user/checkLogin/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "LoginInfo",
      hostname: kongAddress,
      path: "/slark/user/loginInfo/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "QRLogin",
      hostname: kongAddress,
      path: "/slark/qrcode/login/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "SendLoginEmailCode",
      hostname: kongAddress,
      path: "/slark/sendLoginEmailCode/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "LoginByEmailCode",
      hostname: kongAddress,
      path: "/slark/loginByEmailCode/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "UpdateNickname",
      hostname: kongAddress,
      path: "/slark/updateNickname/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "Unregister",
      hostname: kongAddress,
      path: "/slark/unregister/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "RandomNickname",
      hostname: kongAddress,
      path: "/slark/randomNickname/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 0,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
  ]
}

// word endpoints -------------------------------------------------------------------
const wordStatus = {
  name: "word",
  proportion: 0,
  endpoints: [
    {
      name: "FavoriteDefinition",
      hostname: kongAddress,
      path: "/word/user/definition/favorite/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 101040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "FavoritedDefinitions",
      hostname: kongAddress,
      path: "/word/user/definition/favorites/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "FetchAudioURL",
      hostname: kongAddress,
      path: "/word/audio/getAudioURL/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 101040011,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "ProgressBackupStatus",
      hostname: kongAddress,
      path: "/word/user/progress/backup/status/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "UploadProgressBackup",
      hostname: kongAddress,
      path: "/word/user/progress/backup/upload/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
    {
      name: "DownloadProgressBackup",
      hostname: kongAddress,
      path: "/word/user/progress/backup/download/v1",
      method: "POST",
      param: null, 
      respCode: '200',
      respDataCode: 100040102,
      status: false,
      lastDT: 0,
      currentDT: 0
    },
  ]
}

// init emailCount
const initEmailCount = ()=>{
  for (let i = 0; i < kongStatus.endpoints.length; i++) {
    emailCount[kongStatus.endpoints[i].hostname + "" + kongStatus.endpoints[i].path] = 0
  }
  for (let i = 0; i < discourseStatus.endpoints.length; i++) {
    emailCount[discourseStatus.endpoints[i].hostname + "" + discourseStatus.endpoints[i].path] = 0
  }
  for (let i = 0; i < jenkinsStatus.endpoints.length; i++) {
    emailCount[jenkinsStatus.endpoints[i].hostname + "" + jenkinsStatus.endpoints[i].path] = 0
  }
  for (let i = 0; i < slarkUIStatus.endpoints.length; i++) {
    emailCount[slarkUIStatus.endpoints[i].hostname + "" + slarkUIStatus.endpoints[i].path] = 0
  }
  for (let i = 0; i < slarkStatus.endpoints.length; i++) {
    emailCount[slarkStatus.endpoints[i].hostname + "" + slarkStatus.endpoints[i].path] = 0
  }
  for (let i = 0; i < wordStatus.endpoints.length; i++) {
    emailCount[wordStatus.endpoints[i].hostname + "" + wordStatus.endpoints[i].path] = 0
  }
}
initEmailCount()

// interval job
setInterval(() => {
  // kong endpoints
  kongStatus.proportion = 0;
  for (let i = 0; i < kongStatus.endpoints.length; i++) {
    statusHttp2RespCode(
      kongStatus,
      i,
      kongStatus.endpoints[i].hostname, 
      kongStatus.endpoints[i].path,
      kongStatus.endpoints[i].method,
      kongStatus.endpoints[i].param,
      kongStatus.endpoints[i].respCode,
    )
  }

  // discourse endpoints
  discourseStatus.proportion = 0;
  for (let i = 0; i < discourseStatus.endpoints.length; i++) {
    statusHttpRespCode(
      discourseStatus,
      i,
      discourseStatus.endpoints[i].hostname + discourseStatus.endpoints[i].path,
      discourseStatus.endpoints[i].method,
      discourseStatus.endpoints[i].param,
      discourseStatus.endpoints[i].respCode,
    )
  }

  // jenkins endpoints
  jenkinsStatus.proportion = 0;
  for (let i = 0; i < jenkinsStatus.endpoints.length; i++) {
    statusHttpRespCode(
      jenkinsStatus,
      i,
      jenkinsStatus.endpoints[i].hostname + jenkinsStatus.endpoints[i].path,
      jenkinsStatus.endpoints[i].method,
      jenkinsStatus.endpoints[i].param,
      jenkinsStatus.endpoints[i].respCode,
    )
  }

  // slark-ui endpoints
  slarkUIStatus.proportion = 0;
  for (let i = 0; i < slarkUIStatus.endpoints.length; i++) {
    statusHttpRespCode(
      slarkUIStatus,
      i,
      slarkUIStatus.endpoints[i].hostname + slarkUIStatus.endpoints[i].path,
      slarkUIStatus.endpoints[i].method,
      slarkUIStatus.endpoints[i].param,
      slarkUIStatus.endpoints[i].respCode,
    )
  }

  // slark endpoints
  slarkStatus.proportion = 0;
  for (let i = 0; i < slarkStatus.endpoints.length; i++) {
    statusHttp2RespCode(
      slarkStatus,
      i,
      slarkStatus.endpoints[i].hostname, 
      slarkStatus.endpoints[i].path,
      slarkStatus.endpoints[i].method,
      slarkStatus.endpoints[i].param,
      slarkStatus.endpoints[i].respCode,
    );
    if (slarkStatus.endpoints[i].status) {
      slarkStatus.proportion -= 1;
      statusHttp2RespDataCode(
        slarkStatus,
        i,
        slarkStatus.endpoints[i].hostname, 
        slarkStatus.endpoints[i].path,
        slarkStatus.endpoints[i].method,
        slarkStatus.endpoints[i].param,
        slarkStatus.endpoints[i].respDataCode,
      );
    }
  }

  // word endpoints
  wordStatus.proportion = 0;
  for (let i = 0; i < wordStatus.endpoints.length; i++) {
    statusHttp2RespCode(
      wordStatus,
      i,
      wordStatus.endpoints[i].hostname, 
      wordStatus.endpoints[i].path,
      wordStatus.endpoints[i].method,
      wordStatus.endpoints[i].param,
      wordStatus.endpoints[i].respCode,
    );
    if (wordStatus.endpoints[i].status) {
      wordStatus.proportion -= 1;
      statusHttp2RespDataCode(
        wordStatus,
        i,
        wordStatus.endpoints[i].hostname, 
        wordStatus.endpoints[i].path,
        wordStatus.endpoints[i].method,
        wordStatus.endpoints[i].param,
        wordStatus.endpoints[i].respDataCode,
      )
    }
  }
  
}, 1000 * 60 * 10)

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`)
})