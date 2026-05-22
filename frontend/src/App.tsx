import * as React from 'react'

// Импорт страниц приложения
import Trainings from './Trainings'
import Nutrition from './Nutrition'
import Stats from './Stats'
import Calculator from './Calculator'
import Profile from './Profile'
import Admin from './Admin'
import RoleSelector from './RoleSelector'

// Импорт стилей
import './index.css'

// Типы страниц приложения
type Page =
    | 'home'
    | 'trainings'
    | 'nutrition'
    | 'stats'
    | 'calculator'
    | 'profile'
    | 'admin'

// Главный компонент приложения
export default function App() {

    // Текущая страница
    const [page, setPage] =
        React.useState<Page>('home')

    // Данные пользователя
    const [user, setUser] =
        React.useState<any>(null)

    // Проверка admin прав
    const [isAdmin, setIsAdmin] =
        React.useState(false)

    // Выбранная роль
    const [role, setRole] =
        React.useState<'user' | 'admin' | null>(null)

    // Состояние загрузки
    const [loading, setLoading] =
        React.useState(true)

    // Авторизация Telegram Web App
    React.useEffect(() => {

        // Получаем Telegram Mini App API
        const tg = window.Telegram?.WebApp

        // DEV режим для браузера
        if (!tg) {

            console.log(
                'DEV MODE ENABLED',
            )

            // Тестовый пользователь
            setUser({

                id: 1,

                telegram_id: 898030333,

                name: 'Alena Admin',
            })

            // Включаем admin режим
            setIsAdmin(true)

            // Показываем выбор роли
            setRole(null)

            // Отключаем loading
            setLoading(false)

            return
        }

        // Инициализация Telegram Mini App
        tg.ready()

        // Разворачиваем приложение
        tg.expand()

        // Получаем Telegram user
        const telegramUser =
            tg.initDataUnsafe?.user

        console.log(
            'Telegram User:',
            telegramUser,
        )

        // Если пользователь найден
        if (telegramUser) {

            // Отправляем авторизацию на backend
            fetch(
                '/api/auth/telegram',
                {

                    method: 'POST',

                    headers: {
                        'Content-Type': 'application/json',
                    },

                    body: JSON.stringify({

                        telegram_id: telegramUser.id,

                        username: telegramUser.username,

                        first_name: telegramUser.first_name,

                        last_name: telegramUser.last_name,
                    }),
                },
            )
                .then(res => res.json())

                .then(data => {

                    console.log(
                        'AUTH RESPONSE:',
                        data,
                    )

                    // Сохраняем пользователя
                    setUser(data.user)

                    // Сохраняем admin статус
                    setIsAdmin(data.isAdmin)

                    // Отключаем loading
                    setLoading(false)
                })

                .catch(err => {

                    console.error(
                        'AUTH ERROR:',
                        err,
                    )

                    setLoading(false)
                })
        }
        else {

            setLoading(false)
        }

    }, [])

    // Экран загрузки
    if (loading) {

        return (

            <div className="app-shell">

                <div className="card">

                    <h2>
                        Загрузка...
                    </h2>

                </div>

            </div>
        )
    }

    // Экран выбора роли
    if (role === null) {

        return (

            <div className="app-shell">

                <RoleSelector
                    isAdmin={isAdmin}
                    onSelect={setRole}
                />

            </div>
        )
    }

    // Основной интерфейс приложения
    return (

        <div className="app-shell">

            {/* Верхняя часть приложения */}

            <header className="hero">

                {/* Название */}

                <div className="hero-badge">
                    Fitness Bot
                </div>

                {/* Главный заголовок */}

                <h1 className="hero-title">
                    Твой цифровой фитнес-коуч
                </h1>

                {/* Подзаголовок */}

                <p className="hero-subtitle">

                    Тренировки, питание и прогресс —
                    в одном Mini App

                </p>

                {/* Текущий режим */}

                <div className="role-badge">

                    {role === 'admin'
                        ? '👑 Режим администратора'
                        : '👤 Режим пользователя'}

                </div>

            </header>

            {/* Контент страниц */}

            <main className="page-content">

                {/* Главная страница */}

                {page === 'home' && (

                    <HomePage
                        setPage={setPage}
                        isAdmin={isAdmin}
                        role={role}
                    />
                )}

                {/* Тренировки */}

                {page === 'trainings' && (

                    <Trainings
                        setPage={setPage}
                    />
                )}

                {/* Питание */}

                {page === 'nutrition' && (

                    <Nutrition
                        setPage={setPage}
                    />
                )}

                {/* Статистика */}

                {page === 'stats' && (

                    <Stats
                        setPage={setPage}
                    />
                )}

                {/* Профиль */}

                {page === 'profile' && (

                    <Profile
                        setPage={setPage}
                    />
                )}

                {/* Калькулятор */}

                {page === 'calculator' && (

                    <Calculator
                        setPage={setPage}
                    />
                )}

                {/* Admin panel */}

                {page === 'admin'
                    && isAdmin
                    && role === 'admin' && (

                        <Admin />
                    )}

            </main>

            {/* Нижнее меню */}

            <nav className="bottom-nav">

                {/* Главная */}

                <button
                    className={
                        page === 'home'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('home')}
                >

                    <span>🏠</span>
                    <span>Главная</span>

                </button>

                {/* Тренировки */}

                <button
                    className={
                        page === 'trainings'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('trainings')}
                >

                    <span>🏋️</span>
                    <span>Тренировки</span>

                </button>

                {/* Питание */}

                <button
                    className={
                        page === 'nutrition'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('nutrition')}
                >

                    <span>🥗</span>
                    <span>Питание</span>

                </button>

                {/* Статистика */}

                <button
                    className={
                        page === 'stats'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('stats')}
                >

                    <span>📊</span>
                    <span>Статистика</span>

                </button>

                {/* Профиль */}

                <button
                    className={
                        page === 'profile'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('profile')}
                >

                    <span>👤</span>
                    <span>Профиль</span>

                </button>

                {/* Калькулятор */}

                <button
                    className={
                        page === 'calculator'
                            ? 'nav-btn active'
                            : 'nav-btn'
                    }
                    onClick={() => setPage('calculator')}
                >

                    <span>🧮</span>
                    <span>Калькулятор</span>

                </button>

            </nav>

        </div>
    )
}

// Главная страница
function HomePage({
                      setPage,
                      isAdmin,
                      role,
                  }: {
    setPage: (page: Page) => void
    isAdmin: boolean
    role: 'user' | 'admin'
}) {

    return (

        <div className="stack">

            <section className="card">

                {/* Заголовок */}

                <div className="section-head">

                    <h2>
                        Быстрый старт
                    </h2>

                </div>

                {/* Быстрые кнопки */}

                <div className="quick-grid">

                    {/* Кнопка тренировок */}

                    <button
                        className="quick-card"
                        onClick={() => setPage('trainings')}
                    >

                        <div className="quick-icon">
                            🔥
                        </div>

                        <div className="quick-title">
                            Тренировки
                        </div>

                        <div className="quick-desc">
                            Готовые планы и упражнения
                        </div>

                    </button>

                    {/* Кнопка питания */}

                    <button
                        className="quick-card"
                        onClick={() => setPage('nutrition')}
                    >

                        <div className="quick-icon">
                            🥗
                        </div>

                        <div className="quick-title">
                            Питание
                        </div>

                        <div className="quick-desc">
                            Рацион и калории
                        </div>

                    </button>

                    {/* Кнопка статистики */}

                    <button
                        className="quick-card"
                        onClick={() => setPage('stats')}
                    >

                        <div className="quick-icon">
                            📊
                        </div>

                        <div className="quick-title">
                            Статистика
                        </div>

                        <div className="quick-desc">
                            Отслеживание прогресса
                        </div>

                    </button>

                    {/* Кнопка профиля */}

                    <button
                        className="quick-card"
                        onClick={() => setPage('profile')}
                    >

                        <div className="quick-icon">
                            👤
                        </div>

                        <div className="quick-title">
                            Профиль
                        </div>

                        <div className="quick-desc">
                            Данные пользователя
                        </div>

                    </button>

                    {/* Кнопка калькулятора */}

                    <button
                        className="quick-card"
                        onClick={() => setPage('calculator')}
                    >

                        <div className="quick-icon">
                            🧮
                        </div>

                        <div className="quick-title">
                            Калькулятор
                        </div>

                        <div className="quick-desc">
                            ИМТ и калории
                        </div>

                    </button>

                    {/* Admin panel только для admin */}

                    {isAdmin && role === 'admin' && (

                        <button
                            className="quick-card"
                            onClick={() => setPage('admin')}
                        >

                            <div className="quick-icon">
                                👑
                            </div>

                            <div className="quick-title">
                                Admin Panel
                            </div>

                            <div className="quick-desc">
                                Управление платформой
                            </div>

                        </button>
                    )}

                </div>

            </section>

        </div>
    )
}