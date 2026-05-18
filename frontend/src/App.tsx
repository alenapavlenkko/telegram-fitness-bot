import * as React from 'react'
import Trainings from './Trainings'
import Nutrition from './Nutrition'
import Stats from './Stats'
import './index.css'

type Page = 'home' | 'trainings' | 'nutrition' | 'stats' | 'profile'

export default function App() {
    const [page, setPage] = React.useState<Page>('home')

    React.useEffect(() => {
        const tg = window.Telegram?.WebApp
        if (tg) {
            tg.ready()
            tg.expand()
        }
    }, [])

    return (
        <div className="app-shell">
            <header className="hero">
                <div className="hero-badge">Fitness Bot</div>
                <h1 className="hero-title">Твой цифровой фитнес-коуч</h1>
                <p className="hero-subtitle">
                    Тренировки, питание и прогресс — в одном Mini App
                </p>
            </header>

            <main className="page-content">
                {page === 'home' && <HomePage setPage={setPage} />}
                {page === 'trainings' && <Trainings />}
                {page === 'nutrition' && <Nutrition />}
                {page === 'stats' && <Stats />}
                {page === 'profile' && <ProfilePage />}
            </main>

            <nav className="bottom-nav">
                <button className={page === 'home' ? 'nav-btn active' : 'nav-btn'} onClick={() => setPage('home')}>
                    <span>🏠</span>
                    <span>Главная</span>
                </button>

                <button className={page === 'trainings' ? 'nav-btn active' : 'nav-btn'} onClick={() => setPage('trainings')}>
                    <span>🏋️</span>
                    <span>Тренировки</span>
                </button>

                <button className={page === 'nutrition' ? 'nav-btn active' : 'nav-btn'} onClick={() => setPage('nutrition')}>
                    <span>🥗</span>
                    <span>Питание</span>
                </button>

                <button className={page === 'stats' ? 'nav-btn active' : 'nav-btn'} onClick={() => setPage('stats')}>
                    <span>📊</span>
                    <span>Статистика</span>
                </button>

                <button className={page === 'profile' ? 'nav-btn active' : 'nav-btn'} onClick={() => setPage('profile')}>
                    <span>👤</span>
                    <span>Профиль</span>
                </button>
            </nav>
        </div>
    )
}

function HomePage({ setPage }: { setPage: (page: Page) => void }) {
    return (
        <div className="stack">
            <section className="card">
                <div className="section-head">
                    <h2>Быстрый старт</h2>
                </div>

                <div className="quick-grid">
                    <button className="quick-card" onClick={() => setPage('trainings')}>
                        <div className="quick-icon">🔥</div>
                        <div className="quick-title">Тренировки</div>
                        <div className="quick-desc">Готовые планы и упражнения</div>
                    </button>

                    <button className="quick-card" onClick={() => setPage('nutrition')}>
                        <div className="quick-icon">🥗</div>
                        <div className="quick-title">Питание</div>
                        <div className="quick-desc">Рацион, калории и БЖУ</div>
                    </button>

                    <button className="quick-card" onClick={() => setPage('stats')}>
                        <div className="quick-icon">📊</div>
                        <div className="quick-title">Статистика</div>
                        <div className="quick-desc">Общий прогресс</div>
                    </button>

                    <button className="quick-card" onClick={() => setPage('profile')}>
                        <div className="quick-icon">👤</div>
                        <div className="quick-title">Профиль</div>
                        <div className="quick-desc">Данные пользователя</div>
                    </button>
                </div>
            </section>
        </div>
    )
}

function ProfilePage() {
    return (
        <div className="stack">
            <section className="card profile-card">
                <div className="avatar">A</div>
                <div>
                    <h2 className="profile-name">Ваш профиль</h2>
                    <p className="profile-subtitle">Telegram Mini App пользователя</p>
                </div>
            </section>
        </div>
    )
}

declare global {
    interface Window {
        Telegram?: {
            WebApp?: {
                ready: () => void
                expand: () => void
            }
        }
    }
}