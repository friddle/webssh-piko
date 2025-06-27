import request from '@/utils/request'
export function checkSSH(sshInfo) {
    const prefix = "/"+window.SUB_PATH || (process.env.NODE_ENV === 'production' ? '' : '/ws')
    return request.get(`${prefix}/check?sshInfo=${sshInfo}`)
}
