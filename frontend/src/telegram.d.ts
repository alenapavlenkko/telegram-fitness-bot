export {}

declare global {
    interface Window {
        Telegram?: {
            WebApp?: {
                ready: () => void
                expand: () => void
                initData?: string
                initDataUnsafe?: {
                    user?: {
                        id: number
                        first_name?: string
                        last_name?: string
                        username?: string
                    }
                }
            }
        }
    }
}