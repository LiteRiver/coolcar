import camelcaseKeys from 'camelcase-keys'
import { auth } from './proto-gen/auth/auth-pb'

export namespace Coolcar {
  const SERVER_BASE_URL = 'http://localhost:8081'
  const AUTH_ERR = 'AUTH_ERR'

  const authData = {
    token: '',
    expiryMs: 0,
  }

  export interface RequestOption<REQ, RES> {
    method: 'GET' | 'POST' | 'PUT' | 'DELETE'
    path: string
    data?: REQ
    respMarshaller: (r: object) => RES
  }

  export interface AuthOption {
    attachAuthHeader: boolean
    retryOnError: boolean
  }

  export async function sendRequestWithAuthRetry<REQ, RES>(
    o: RequestOption<REQ, RES>,
    authOpt?: AuthOption
  ): Promise<RES> {
    const aOpt = authOpt || {
      attachAuthHeader: true,
      retryOnError: true,
    }

    try {
      await login()
      return sendRequest(o, aOpt)
    } catch (err) {
      if (err === AUTH_ERR && aOpt.retryOnError) {
        authData.token = ''
        authData.expiryMs = 0
        return sendRequestWithAuthRetry(o, {
          ...aOpt,
          retryOnError: false,
        })
      } else {
        throw err
      }
    }
  }

  export async function login() {
    if (authData.token && authData.expiryMs >= Date.now()) return

    const wxRes = await wxLogin()
    const reqMs = Date.now()
    const res = await sendRequest<auth.v1.ILoginRequest, auth.v1.ILoginResponse>(
      {
        path: '/v1/auth/login',
        method: 'POST',
        data: {
          code: wxRes.code,
        },
        respMarshaller: auth.v1.LoginResponse.fromObject,
      },
      {
        attachAuthHeader: false,
        retryOnError: false,
      }
    )

    authData.token = res.accessToken!
    authData.expiryMs = reqMs + res.expiresIn! * 1000
  }

  function sendRequest<REQ, RES>(
    o: RequestOption<REQ, RES>,
    authOpt: AuthOption
  ): Promise<RES> {
    return new Promise((resolve, reject) => {
      const header: Record<string, any> = {}

      if (authOpt.attachAuthHeader) {
        if (authData.token && authData.expiryMs >= Date.now()) {
          header.authorization = `Bearer ${authData.token}`
        } else {
          reject(AUTH_ERR)
          return
        }
      }

      wx.request({
        url: `${SERVER_BASE_URL}${o.path}`,
        method: o.method,
        data: o.data,
        header,
        success: (res) => {
          if (res.statusCode === 401) {
            reject(AUTH_ERR)
          } else if (res.statusCode >= 400) {
            reject(res)
          } else {
            resolve(o.respMarshaller(camelcaseKeys(res.data as object, { deep: true })))
          }
        },
        fail: reject,
      })
    })
  }

  function wxLogin(): Promise<WechatMiniprogram.LoginSuccessCallbackResult> {
    return new Promise((resolve, reject) => {
      return wx.login({
        success: resolve,
        fail: reject,
      })
    })
  }

  export interface UploadFileOpts {
    localPath: string
    url: string
  }
  export async function uploadFile(opts: UploadFileOpts) {
    const data = wx.getFileSystemManager().readFileSync(opts.localPath)
    return new Promise((resolve, reject) => {
      wx.request({
        method: 'PUT',
        url: opts.url,
        data,
        header: {
          'Content-Type': 'application/octet-stream',
        },
        success: (res) => {
          if (res.statusCode >= 400) {
            reject(res)
          } else {
            resolve(undefined)
          }
        },
        fail: reject,
      })
    })
  }
}
