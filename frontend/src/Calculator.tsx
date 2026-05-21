import * as React from 'react'

type Gender = 'female' | 'male' | 'other'
type Goal = 'Похудение' | 'Набор массы' | 'Поддержание формы'
type Activity = 'Низкая' | 'Средняя' | 'Высокая'

export default function Calculator() {
    const [weight, setWeight] = React.useState(55)
    const [height, setHeight] = React.useState(165)
    const [age, setAge] = React.useState(20)
    const [gender, setGender] = React.useState<Gender>('female')
    const [activity, setActivity] = React.useState<Activity>('Средняя')
    const [goal, setGoal] = React.useState<Goal>('Похудение')

    const heightM = height / 100
    const bmi = weight > 0 && height > 0 ? weight / (heightM * heightM) : 0

    const genderValue = gender === 'male' ? 5 : -161
    const bmr = 10 * weight + 6.25 * height - 5 * age + genderValue

    const activityMultiplier =
        activity === 'Низкая' ? 1.2 :
            activity === 'Средняя' ? 1.55 :
                1.725

    const maintenanceCalories = Math.round(bmr * activityMultiplier)

    const targetCalories =
        goal === 'Похудение'
            ? maintenanceCalories - 300
            : goal === 'Набор массы'
                ? maintenanceCalories + 300
                : maintenanceCalories

    const bmiStatus =
        bmi < 18.5
            ? 'Недостаточный вес'
            : bmi < 25
                ? 'Норма'
                : bmi < 30
                    ? 'Избыточный вес'
                    : 'Ожирение'

    return (
        <div className="stack">
            <section className="card">
                <div className="section-head">
                    <h2>Калькулятор</h2>
                    <span className="pill">BMI</span>
                </div>

                <div className="form-grid">
                    <label className="form-field">
                        <span>Вес, кг</span>
                        <input
                            type="number"
                            value={weight}
                            onChange={(e) => setWeight(Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Рост, см</span>
                        <input
                            type="number"
                            value={height}
                            onChange={(e) => setHeight(Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Возраст</span>
                        <input
                            type="number"
                            value={age}
                            onChange={(e) => setAge(Number(e.target.value))}
                        />
                    </label>

                    <label className="form-field">
                        <span>Пол</span>
                        <select value={gender} onChange={(e) => setGender(e.target.value as Gender)}>
                            <option value="female">Женский</option>
                            <option value="male">Мужской</option>
                            <option value="other">Другое</option>
                        </select>
                    </label>

                    <label className="form-field">
                        <span>Активность</span>
                        <select value={activity} onChange={(e) => setActivity(e.target.value as Activity)}>
                            <option>Низкая</option>
                            <option>Средняя</option>
                            <option>Высокая</option>
                        </select>
                    </label>

                    <label className="form-field">
                        <span>Цель</span>
                        <select value={goal} onChange={(e) => setGoal(e.target.value as Goal)}>
                            <option>Похудение</option>
                            <option>Набор массы</option>
                            <option>Поддержание формы</option>
                        </select>
                    </label>
                </div>
            </section>

            <section className="card">
                <div className="section-head">
                    <h2>Результат</h2>
                </div>

                <div className="stats-card">
                    <div className="stat-item">
                        <div className="stat-label">ИМТ</div>
                        <div className="stat-value">{bmi ? bmi.toFixed(1) : '—'}</div>
                        <div className="stat-hint">{bmiStatus}</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Базовый обмен</div>
                        <div className="stat-value">{Math.round(bmr)}</div>
                        <div className="stat-hint">ккал/день</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Поддержание</div>
                        <div className="stat-value">{maintenanceCalories}</div>
                        <div className="stat-hint">ккал/день</div>
                    </div>

                    <div className="stat-item">
                        <div className="stat-label">Для цели</div>
                        <div className="stat-value">{targetCalories}</div>
                        <div className="stat-hint">ккал/день</div>
                    </div>
                </div>
            </section>
        </div>
    )
}