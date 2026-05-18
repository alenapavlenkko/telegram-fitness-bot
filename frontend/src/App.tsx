import * as React from 'react'
import Trainings from './Trainings'
import './index.css'

type Page = 'home' | 'trainings' | 'profile'

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
                {page === 'profile' && <ProfilePage />}
            </main>

            <nav className="bottom-nav">
                <button
                    className={page === 'home' ? 'nav-btn active' : 'nav-btn'}
                    onClick={() => setPage('home')}
                >
                    <span>🏠</span>
                    <span>Главная</span>
                </button>

                <button
                    className={page === 'trainings' ? 'nav-btn active' : 'nav-btn'}
                    onClick={() => setPage('trainings')}
                >
                    <span>🏋️</span>
                    <span>Тренировки</span>
                </button>

                <button
                    className={page === 'profile' ? 'nav-btn active' : 'nav-btn'}
                    onClick={() => setPage('profile')}
                >
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
            <section className="card stats-card">
                <div className="stat-item">
                    <div className="stat-label">Сегодня</div>
                    <div className="stat-value">4/5</div>
                    <div className="stat-hint">целей закрыто</div>
                </div>
                <div className="stat-item">
                    <div className="stat-label">Активность</div>
                    <div className="stat-value">48 мин</div>
                    <div className="stat-hint">тренировок</div>
                </div>
            </section>

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

                    <button className="quick-card">
                        <div className="quick-icon">🥗</div>
                        <div className="quick-title">Питание</div>
                        <div className="quick-desc">Рацион и калории</div>
                    </button>

                    <button className="quick-card">
                        <div className="quick-icon">⚖️</div>
                        <div className="quick-title">Вес</div>
                        <div className="quick-desc">Отслеживание прогресса</div>
                    </button>

                    <button className="quick-card">
                        <div className="quick-icon">📈</div>
                        <div className="quick-title">Статистика</div>
                        <div className="quick-desc">Результаты и динамика</div>
                    </button>
                </div>
            </section>

            <section className="card highlight-card">
                <div>
                    <div className="highlight-label">Мотивация дня</div>
                    <div className="highlight-title">
                        Маленькие шаги каждый день дают большой результат
                    </div>
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

            <section className="card">
                <div className="profile-row">
                    <span>Статус</span>
                    <strong>Активен</strong>
                </div>
                <div className="profile-row">
                    <span>План</span>
                    <strong>Базовый</strong>
                </div>
                <div className="profile-row">
                    <span>Прогресс</span>
                    <strong>Хороший</strong>
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