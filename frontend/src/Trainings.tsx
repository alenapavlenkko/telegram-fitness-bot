import * as React from 'react'

type Training = {
    ID?: number
    id?: number
    Title?: string
    title?: string
    Description?: string
    description?: string
    Difficulty?: string
    difficulty?: string
    Duration?: number
    duration?: number
    YouTubeLink?: string
    youtubeLink?: string
}

export default function Trainings() {
    const [data, setData] = React.useState<Training[]>([])
    const [loading, setLoading] = React.useState(true)
    const [error, setError] = React.useState<string | null>(null)

    React.useEffect(() => {
        fetch('/api/trainings')
            .then(async (res) => {
                if (!res.ok) {
                    throw new Error(`Ошибка API: ${res.status}`)
                }
                return res.json()
            })
            .then(setData)
            .catch((err: Error) => setError(err.message))
            .finally(() => setLoading(false))
    }, [])

    if (loading) {
        return (
            <div className="stack">
                <section className="card">
                    <h2>Тренировки</h2>
                    <p className="muted">Загрузка тренировок...</p>
                </section>
            </div>
        )
    }

    if (error) {
        return (
            <div className="stack">
                <section className="card">
                    <h2>Тренировки</h2>
                    <p className="error-text">{error}</p>
                </section>
            </div>
        )
    }

    return (
        <div className="stack">
            <section className="card">
                <div className="section-head">
                    <h2>Доступные тренировки</h2>
                    <span className="pill">{data.length}</span>
                </div>

                {data.length === 0 ? (
                    <p className="muted">Тренировок пока нет</p>
                ) : (
                    <div className="training-list">
                        {data.map((t, i) => {
                            const id = t.ID ?? t.id ?? i
                            const title = t.Title ?? t.title ?? 'Без названия'
                            const description = t.Description ?? t.description ?? ''
                            const duration = t.Duration ?? t.duration ?? 0
                            const difficulty = t.Difficulty ?? t.difficulty ?? 'Не указано'
                            const video = t.YouTubeLink ?? t.youtubeLink ?? ''

                            return (
                                <article key={id} className="training-card">
                                    <div className="training-top">
                                        <div>
                                            <h3 className="training-title">{title}</h3>
                                            <p className="training-duration">
                                                ⏱ {duration} мин · {difficulty}
                                            </p>
                                        </div>
                                        <div className="training-badge">Workout</div>
                                    </div>

                                    {description && (
                                        <p className="training-description">{description}</p>
                                    )}

                                    <div className="training-actions">
                                        {video ? (
                                            <a
                                                className="primary-btn"
                                                href={video}
                                                target="_blank"
                                                rel="noreferrer"
                                            >
                                                Смотреть видео
                                            </a>
                                        ) : (
                                            <button className="primary-btn" disabled>
                                                Видео скоро
                                            </button>
                                        )}
                                    </div>
                                </article>
                            )
                        })}
                    </div>
                )}
            </section>
        </div>
    )
}