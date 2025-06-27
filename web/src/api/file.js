import request from '@/utils/request'

// 获取API前缀的公用方法
function getApiPrefix() {
    return "/" + (window.SUB_PATH || (process.env.NODE_ENV === 'production' ? '' : '/ws'))
}

// 文件列表接口
export function fileList(path, sshInfo) {
    const prefix = getApiPrefix()
    return request.get(`${prefix}/file/list?path=${path}&sshInfo=${sshInfo}`, {
        headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

// 文件上传接口
export function fileUpload(uploadData, file) {
    const prefix = process.env.NODE_ENV === 'production' ? `${location.origin}${getApiPrefix()}` : `api${getApiPrefix()}`
    return request.post(`${prefix}/file/upload`, uploadData, {
        headers: {
            'Content-Type': 'multipart/form-data',
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

// 文件下载接口
export function fileDownload(path, sshInfo) {
    const prefix = process.env.NODE_ENV === 'production' ? `${location.origin}${getApiPrefix()}` : `api${getApiPrefix()}`
    return request.get(`${prefix}/file/download?path=${path}&sshInfo=${sshInfo}`, {
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        },
        responseType: 'blob'
    })
}

// 文件上传进度查询接口
export function fileProgress(id) {
    const prefix = getApiPrefix()
    return request.get(`${prefix}/file/progress?id=${id}`, {
        headers: {
            'X-Requested-With': 'XMLHttpRequest'
        }
    })
}

// 获取上传URL
export function getUploadUrl() {
    return `${process.env.NODE_ENV === 'production' ? `${location.origin}${getApiPrefix()}` : `api${getApiPrefix()}`}/file/upload`
}

// 创建WebSocket连接用于文件上传进度查询
export function createFileProgressWebSocket(fileId, onMessage, onClose, onError) {
    const ws = new WebSocket(`${(location.protocol === 'http:' ? 'ws' : 'wss')}://${location.host}${getApiPrefix()}/file/progress?id=${fileId}`)
    
    if (onMessage) {
        ws.onmessage = onMessage
    }
    
    if (onClose) {
        ws.onclose = onClose
    }
    
    if (onError) {
        ws.onerror = onError
    }
    
    return ws
}

// 处理文件下载
export async function handleFileDownload(path, sshInfo) {
    try {
        const response = await fileDownload(path, sshInfo)
        // 创建下载链接
        const blob = new Blob([response])
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = path.split('/').pop()
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
        return { success: true }
    } catch (error) {
        console.error('下载文件错误:', error)
        return { success: false, error }
    }
}
