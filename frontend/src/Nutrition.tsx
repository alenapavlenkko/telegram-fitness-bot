import * as React from 'react'

type Nutrition = {
    ID?: number
    id?: number
    Title?: string
    title?: string
    Description?: string
    description?: string
    Calories?: number
    calories?: number
    Protein?: number
    protein?: number
    Carbs?: number
    carbs?: number
    Fats?: number
    fats?: number
}

export default function Nutrition({
                                      setPage,
                                  }: {
    setPage: (page: any) => void
}) {
    const [data, setData] = React.useState<Nutrition[]>([])
    const [loading, setLoading] = React.useState(true)
    const [error, setError] = React.useState<string | null>(null)

    React.useEffect(() => {
        fetch('/api/nutrition')
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
                    <h2>Питание</h2>
                    <p className="muted">Загрузка питания...</p>
                </section>
            </div>
        )
    }

    if (error) {
        return (
            <div className="stack">
                <section className="card">
                    <h2>Питание</h2>
                    <p className="error-text">{error}</p>
                </section>
            </div>
        )
    }

    return (
        <div className="stack">
            <section className="card">
                <div className="section-head">
                    <h2>Питание</h2>
                    <span className="pill">{data.length}</span>
                </div>

                {data.length === 0 ? (
                    <p className="muted">Блюд пока нет</p>
                ) : (
                    <div className="training-list">
                        {data.map((n, i) => {
                            const title = n.Title ?? n.title ?? 'Без названия'
                            const description = n.Description ?? n.description ?? ''
                            const calories = n.Calories ?? n.calories ?? 0
                            const protein = n.Protein ?? n.protein ?? 0
                            const carbs = n.Carbs ?? n.carbs ?? 0
                            const fats = n.Fats ?? n.fats ?? 0

                            return (
                                <article
                                    className="training-card"
                                    key={n.ID ?? n.id ?? i}
                                >
                                    <div className="training-top">
                                        <div>
                                            <h3 className="training-title">
                                                {title}
                                            </h3>

                                            <p className="training-duration">
                                                🔥 {calories} ккал
                                            </p>
                                        </div>

                                        <div className="training-badge">
                                            🍎 Food
                                        </div>
                                    </div>

                                    {description && (
                                        <p className="training-description">
                                            {description}
                                        </p>
                                    )}

                                    <div
                                        className="stats-card"
                                        style={{ marginTop: 12 }}
                                    >
                                        <div className="stat-item">
                                            <div className="stat-label">
                                                Белки
                                            </div>

                                            <div className="stat-value">
                                                {protein}г
                                            </div>
                                        </div>

                                        <div className="stat-item">
                                            <div className="stat-label">
                                                Углеводы
                                            </div>

                                            <div className="stat-value">
                                                {carbs}г
                                            </div>
                                        </div>

                                        <div className="stat-item">
                                            <div className="stat-label">
                                                Жиры
                                            </div>

                                            <div className="stat-value">
                                                {fats}г
                                            </div>
                                        </div>
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