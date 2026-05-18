import * as React from 'react'

type Training = {
    id?: number
    title: string
    description?: string
    duration: number
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
                        {data.map((t, i) => (
                            <article key={t.id ?? i} className="training-card">
                                <div className="training-top">
                                    <div>
                                        <h3 className="training-title">{t.title}</h3>
                                        <p className="training-duration">⏱ {t.duration} мин</p>
                                    </div>
                                    <div className="training-badge">Workout</div>
                                </div>

                                {t.description && (
                                    <p className="training-description">{t.description}</p>
                                )}

                                <div className="training-actions">
                                    {t.youtubeLink ? (
                                        <a
                                            className="primary-btn"
                                            href={t.youtubeLink}
                                            target="_blank"
                                            rel="noreferrer"
                                        >
                                            Смотреть
                                        </a>
                                    ) : (
                                        <button className="primary-btn" disabled>
                                            Видео скоро
                                        </button>
                                    )}
                                </div>
                            </article>
                        ))}
                    </div>
                )}
            </section>
        </div>
    )
}