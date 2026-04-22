declare namespace WechatMiniprogram {
  interface IAnyObject {
    [key: string]: any
  }

  interface PageOptions<D extends IAnyObject, M extends IAnyObject> {
    data: D
    [key: string]: any
  }

  interface App.Instance<D extends IAnyObject> {
    globalData: D
    onLaunch?: () => void
    onShow?: () => void
    onHide?: () => void
  }

  interface TouchEvent {
    type: string
    currentTarget: {
      dataset: IAnyObject
    }
    target: {
      dataset: IAnyObject
    }
    detail: IAnyObject
    touches: IAnyObject[]
    changedTouches: IAnyObject[]
    timeStamp: number
  }

  interface InputEvent {
    type: string
    detail: {
      value: string
      cursor?: number
      keyCode?: number
    }
  }

  interface PickerChange {
    type: string
    detail: {
      value: number | number[]
    }
  }

  interface RequestSuccessCallbackResult {
    data: string | IAnyObject | ArrayBuffer
    statusCode: number
    header: IAnyObject
    cookies: string[]
    profile: IAnyObject
  }
}

interface IAppOption {
  globalData: {
    userInfo: WechatMiniprogram.IAnyObject | null
    token: string
    baseUrl: string
  }
  onLaunch?: () => void
}

declare function Page(options: WechatMiniprogram.IAnyObject): void
declare function Component(options: WechatMiniprogram.IAnyObject): void
declare function getApp(): WechatMiniprogram.App.Instance<WechatMiniprogram.IAnyObject>