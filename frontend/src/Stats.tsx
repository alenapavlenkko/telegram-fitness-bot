import * as React from 'react'

type StatsData = {
    trainingsCount: number
    nutritionCount: number
    totalMinutes: number
    totalCalories: number
}

export default function Stats() {
    const [stats, setStats] = React.useState<StatsData | null>(null)

    React.useEffect(() => {
        fetch('/api/stats')
            .then(res => res.json())
            .then(setStats)
    }, [])

    if (!stats) return <div className="card">Загрузка статистики...</div>

    return (
        <div className="stack">
            <section className="card">
                <h2>Статистика</h2>

                <div className="stats-card">
                    <div className="stat-item">
                        <div className="stat-label">Тренировки</div>
                        <div className="stat-value">{stats.trainingsCount}</div>
                        <div className="stat-hint">в базе</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Питание</div>
                        <div className="stat-value">{stats.nutritionCount}</div>
                        <div className="stat-hint">блюд</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Минуты</div>
                        <div className="stat-value">{stats.totalMinutes}</div>
                        <div className="stat-hint">тренировок</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Калории</div>
                        <div className="stat-value">{stats.totalCalories}</div>
                        <div className="stat-hint">в рационе</div>
                    </div>
                </div>
            </section>
        </div>
    )
}