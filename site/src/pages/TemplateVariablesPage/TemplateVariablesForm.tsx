import TextField from "@material-ui/core/TextField"
import {
  CreateTemplateVersionRequest,
  TemplateVersion,
  TemplateVersionVariable,
  VariableValue,
} from "api/typesGenerated"
import { FormikContextType, FormikTouched, useFormik } from "formik"
import { FC } from "react"
import { getFormHelpers } from "util/formUtils"
import * as Yup from "yup"
import { useTranslation } from "react-i18next"
import {
  FormFields,
  FormSection,
  HorizontalForm,
  FormFooter,
} from "components/HorizontalForm/HorizontalForm"

export const getValidationSchema = (): Yup.AnyObjectSchema => Yup.object()

export interface TemplateVariablesForm {
  templateVersion: TemplateVersion
  templateVariables: TemplateVersionVariable[]
  onSubmit: (data: CreateTemplateVersionRequest) => void
  onCancel: () => void
  isSubmitting: boolean
  error?: unknown
  // Helpful to show field errors on Storybook
  initialTouched?: FormikTouched<CreateTemplateVersionRequest>
}

export const TemplateVariablesForm: FC<TemplateVariablesForm> = ({
  templateVersion,
  templateVariables,
  onSubmit,
  onCancel,
  error,
  isSubmitting,
  initialTouched,
}) => {
  const validationSchema = getValidationSchema()
  const form: FormikContextType<CreateTemplateVersionRequest> =
    useFormik<CreateTemplateVersionRequest>({
      initialValues: {
        template_id: templateVersion.template_id,
        provisioner: "terraform",
        storage_method: "file",
        tags: {},
        // FIXME file_id: null,
        user_variable_values:
          selectInitialUserVariableValues(templateVariables),
      },
      validationSchema,
      onSubmit: onSubmit,
      initialTouched,
    })
  const getFieldHelpers = getFormHelpers<CreateTemplateVersionRequest>(
    form,
    error,
  )
  const { t } = useTranslation("templateVariablesPage")

  return (
    <HorizontalForm
      onSubmit={form.handleSubmit}
      aria-label={t("formAriaLabel")}
    >
      {templateVariables.map((templateVariable, index) => (
        <FormSection
          key={templateVariable.name}
          title={templateVariable.name}
          description={templateVariable.description}
        >
          <FormFields>
            <TextField
              {...getFieldHelpers("user_variable_values[" + index + "].value")}
              disabled={isSubmitting}
              autoFocus
              fullWidth
              label={templateVariable.name}
              variant="outlined"
            />
          </FormFields>
        </FormSection>
      ))}

      <FormFooter onCancel={onCancel} isLoading={isSubmitting} />
    </HorizontalForm>
  )
}

export const selectInitialUserVariableValues = (
  templateVariables: TemplateVersionVariable[],
): VariableValue[] => {
  const defaults: VariableValue[] = []
  templateVariables.forEach((templateVariable) => {
    if (templateVariable.sensitive) {
      defaults.push({
        name: templateVariable.name,
        value: "",
      })
      return
    }

    if (
      templateVariable.value === "" &&
      templateVariable.default_value !== ""
    ) {
      defaults.push({
        name: templateVariable.name,
        value: templateVariable.default_value,
      })
      return
    }

    defaults.push({
      name: templateVariable.name,
      value: templateVariable.value,
    })
  })
  return defaults
}
