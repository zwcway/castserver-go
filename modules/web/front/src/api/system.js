import { socket } from '@/common/request'

export function config() {
    return socket.send('status', { sct: 'config' })
}